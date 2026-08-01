package mautic

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// BasicAuth authenticates with a Mautic user's username and password.
//
// Mautic has supported this since 2.3.0, and it must be enabled by the
// administrator under Configuration -> API Settings. It is the simplest thing
// that works and the right default for scripts and agents.
type BasicAuth struct {
	Username string
	Password string
}

// Apply sets the Authorization header.
func (a *BasicAuth) Apply(req *http.Request) error {
	if a.Username == "" || a.Password == "" {
		return fmt.Errorf("basic auth needs both a username and a password")
	}
	token := base64.StdEncoding.EncodeToString([]byte(a.Username + ":" + a.Password))
	req.Header.Set("Authorization", "Basic "+token)
	return nil
}

// Describe returns a secret-free description.
func (a *BasicAuth) Describe() string { return "basic auth as " + a.Username }

// TokenPersistFunc stores a refreshed token set. The OAuth2 authenticator calls
// it whenever x/oauth2 rotates the tokens, because Mautic issues a *new*
// refresh token on every refresh and losing it means re-authorising by hand.
type TokenPersistFunc func(accessToken, refreshToken string, expiry time.Time) error

// OAuth2Auth authenticates with a bearer token, refreshing it silently.
type OAuth2Auth struct {
	source  oauth2.TokenSource
	persist TokenPersistFunc
	client  string

	mu   sync.Mutex
	last oauth2.Token
}

// OAuth2Config describes the instance's OAuth2 endpoints and the client
// registered under Settings -> API Credentials.
func OAuth2Config(baseURL, clientID, clientSecret, redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseURL + "/oauth/v2/authorize",
			TokenURL: baseURL + "/oauth/v2/token",
			// Mautic expects the client credentials in the POST body.
			AuthStyle: oauth2.AuthStyleInParams,
		},
		// As of 1.1.2 Mautic does not use OAuth2 scopes, so requesting any
		// would be rejected rather than ignored.
		Scopes: nil,
	}
}

// NewOAuth2Auth builds an authenticator from a stored token set.
//
// When the client ID and secret are present the token is refreshed
// automatically as it expires; with only an access token it is used as-is,
// which is what a CI runner holding MAUTIC_ACCESS_TOKEN wants.
func NewOAuth2Auth(ctx context.Context, cfg *oauth2.Config, tok *oauth2.Token, persist TokenPersistFunc) *OAuth2Auth {
	a := &OAuth2Auth{persist: persist, client: cfg.ClientID}
	if cfg.ClientID != "" && cfg.ClientSecret != "" && tok.RefreshToken != "" {
		a.source = oauth2.ReuseTokenSource(tok, cfg.TokenSource(ctx, tok))
	} else {
		a.source = oauth2.StaticTokenSource(tok)
	}
	if tok != nil {
		a.last = *tok
	}
	return a
}

// Apply sets the bearer header, refreshing the token first when needed.
func (a *OAuth2Auth) Apply(req *http.Request) error {
	tok, err := a.source.Token()
	if err != nil {
		return fmt.Errorf("could not obtain an OAuth2 access token: %w\n  Mautic refresh tokens expire (14 days by default) — run `mautic auth login` to re-authorise", err)
	}
	a.remember(tok)
	tok.SetAuthHeader(req)
	return nil
}

// remember persists a token set the moment it changes.
func (a *OAuth2Auth) remember(tok *oauth2.Token) {
	if a.persist == nil || tok == nil {
		return
	}
	a.mu.Lock()
	changed := tok.AccessToken != a.last.AccessToken || tok.RefreshToken != a.last.RefreshToken
	if changed {
		a.last = *tok
	}
	a.mu.Unlock()
	if !changed {
		return
	}
	// A failure to persist must not break the request in flight; the token
	// still works for this run, it just will not survive it.
	_ = a.persist(tok.AccessToken, tok.RefreshToken, tok.Expiry)
}

// Describe returns a secret-free description.
func (a *OAuth2Auth) Describe() string {
	if a.client == "" {
		return "oauth2 bearer token"
	}
	return "oauth2 client " + a.client
}
