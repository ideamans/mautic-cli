package mautic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// DefaultRedirectURI is the callback this CLI listens on by default. It has to
// be registered verbatim in Mautic under Settings -> API Credentials, so the
// port cannot be random — change it with --redirect-uri and register the same
// value on both sides.
const DefaultRedirectURI = "http://localhost:8080/oauth/callback"

// LoginOptions configures the interactive authorization code flow.
type LoginOptions struct {
	Config *oauth2.Config
	// OpenBrowser attempts to open the authorize URL. When it fails the URL is
	// printed instead of erroring: headless sessions are normal.
	OpenBrowser bool
	// Timeout bounds how long to wait for the user. Zero means 5 minutes.
	Timeout time.Duration
	// Out receives progress messages. Must be stderr, never stdout.
	Out io.Writer
}

// Login runs the OAuth2 authorization code flow against a local callback
// server and returns the resulting tokens.
//
// Mautic supports only the authorization_code, refresh_token and
// client_credentials grants, and does not use scopes.
func Login(ctx context.Context, o LoginOptions) (*oauth2.Token, error) {
	if o.Config.ClientID == "" || o.Config.ClientSecret == "" {
		return nil, fmt.Errorf("OAuth2 needs a client ID and secret; create one in Mautic under Settings -> API Credentials and pass --client-id/--client-secret")
	}
	redirect := o.Config.RedirectURL
	if redirect == "" {
		redirect = DefaultRedirectURI
		o.Config.RedirectURL = redirect
	}
	addr, path, err := listenTarget(redirect)
	if err != nil {
		return nil, err
	}
	timeout := o.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	out := o.Out
	if out == nil {
		out = io.Discard
	}

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("could not generate a state value: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not listen on %s for the OAuth2 callback: %w\n  something else is using the port, or the redirect URI names a host this machine is not", addr, err)
	}
	defer listener.Close()

	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			msg := strings.TrimSpace(e + " " + q.Get("error_description"))
			http.Error(w, "OAuth2 error: "+msg, http.StatusBadRequest)
			results <- result{err: fmt.Errorf("Mautic refused the authorization: %s", msg)}
			return
		}
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			results <- result{err: fmt.Errorf("the state parameter did not match; the callback may not have come from this login attempt")}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "no authorization code", http.StatusBadRequest)
			results <- result{err: fmt.Errorf("Mautic redirected back without an authorization code")}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><meta charset="utf-8"><title>mautic-cli</title>`+
			`<body style="font-family:system-ui;padding:3rem"><h1>Authorised</h1>`+
			`<p>Tokens have been saved. You can close this tab and return to the terminal.</p></body>`)
		results <- result{code: code}
	})

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	authURL := o.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Fprintf(out, "Listening for the OAuth2 callback on %s\n", redirect)
	fmt.Fprintf(out, "Authorize at:\n  %s\n", authURL)
	if o.OpenBrowser {
		if err := openBrowser(authURL); err != nil {
			fmt.Fprintf(out, "Could not open a browser (%v) — open the URL above by hand.\n", err)
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out after %s waiting for the callback on %s", timeout, redirect)
	case res := <-results:
		if res.err != nil {
			return nil, res.err
		}
		tok, err := o.Config.Exchange(ctx, res.code)
		if err != nil {
			return nil, fmt.Errorf("could not exchange the authorization code for a token: %w\n  the redirect URI must match the one registered in Mautic exactly, including the port and path", err)
		}
		return tok, nil
	}
}

// ClientCredentials obtains a token with the client_credentials grant, which
// needs no browser and no user. Use it wherever the Mautic instance allows it:
// it is the only OAuth2 path that works unattended.
func ClientCredentials(ctx context.Context, baseURL, clientID, clientSecret string) (*oauth2.Token, error) {
	cfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     baseURL + "/oauth/v2/token",
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	tok, err := cfg.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("the client_credentials grant failed: %w\n  not every Mautic instance enables it; fall back to `mautic auth login` or to basic auth", err)
	}
	return tok, nil
}

// listenTarget turns a redirect URI into a listen address and a handler path.
func listenTarget(redirect string) (addr, path string, err error) {
	u, err := url.Parse(redirect)
	if err != nil {
		return "", "", fmt.Errorf("could not parse the redirect URI %q: %w", redirect, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", fmt.Errorf("the redirect URI must be http or https, got %q", redirect)
	}
	port := u.Port()
	if port == "" {
		port = "80"
	}
	path = u.Path
	if path == "" {
		path = "/"
	}
	return "127.0.0.1:" + port, path, nil
}

// openBrowser is platform specific and allowed to fail.
func openBrowser(target string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, append(args, target)...).Start()
}
