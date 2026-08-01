package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ideamans/mautic-cli/internal/config"
	"github.com/ideamans/mautic-cli/internal/mautic"
)

func newAuthCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "auth",
		Short: "Inspect and obtain credentials",
	}
	group.AddCommand(newAuthStatusCommand(), newAuthLoginCommand(), newAuthLogoutCommand())
	return group
}

// newAuthStatusCommand is the first thing every usage skill runs. It must work
// with nothing configured and say precisely what is missing — so it never
// returns the resolution error, it reports it.
func newAuthStatusCommand() *cobra.Command {
	var verify bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show which credentials would be used, and whether they work",
		Long: `Show which credentials would be used, and whether they work.

Run this before anything that writes. It never prints a secret; it prints where
each value came from, so that "which instance am I about to change" has an
answer.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()

			cfg, cfgErr := config.Load()
			path, _ := config.FilePath()
			fmt.Fprintf(out, "config file:   %s\n", path)
			if cfgErr != nil {
				fmt.Fprintf(out, "               (could not be read: %v)\n", cfgErr)
			} else if len(cfg.Profiles) == 0 {
				fmt.Fprintf(out, "profiles:      (none)\n")
			} else {
				fmt.Fprintf(out, "profiles:      %v\n", cfg.ProfileNames())
				fmt.Fprintf(out, "default:       %s\n", orNone(cfg.DefaultProfile))
			}

			creds, err := resolve()
			if err != nil {
				fmt.Fprintf(out, "credentials:   NOT USABLE\n\n%v\n", err)
				return errSilent{}
			}

			fmt.Fprintf(out, "instance:      %s\n", creds.BaseURL)
			fmt.Fprintf(out, "api endpoint:  %s\n", creds.APIURL())
			fmt.Fprintf(out, "auth method:   %s\n", creds.Method)
			fmt.Fprintf(out, "identity:      %s\n", identityOf(creds))
			fmt.Fprintf(out, "source:        %s\n", creds.Source)
			if creds.Method == config.AuthOAuth2 && !creds.Expiry.IsZero() {
				remaining := time.Until(creds.Expiry).Round(time.Second)
				if remaining > 0 {
					fmt.Fprintf(out, "token expires: %s (in %s)\n", creds.Expiry.Format(time.RFC3339), remaining)
				} else {
					fmt.Fprintf(out, "token expires: %s (expired; it will be refreshed on the next call)\n", creds.Expiry.Format(time.RFC3339))
				}
			}

			if !verify {
				fmt.Fprintf(out, "\nCredentials are complete. Add --verify to call the API and confirm they work.\n")
				return nil
			}

			client, _, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			raw, err := client.Get(cmd.Context(), "/users/self", nil)
			if err != nil {
				fmt.Fprintf(out, "\nverification:  FAILED\n\n%v\n", err)
				return errSilent{}
			}
			// /users/self returns the user object directly, unlike every other
			// single-record endpoint, which wraps it in {"user": {...}}. Accept
			// both so this keeps working if that is ever made consistent.
			type selfUser struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Role     struct {
					Name    string `json:"name"`
					IsAdmin bool   `json:"isAdmin"`
				} `json:"role"`
			}
			var envelope struct {
				User *selfUser `json:"user"`
			}
			var user selfUser
			if err := unmarshalStrict(raw, &envelope); err == nil && envelope.User != nil {
				user = *envelope.User
			} else {
				_ = unmarshalStrict(raw, &user)
			}

			if user.Username != "" {
				admin := ""
				if user.Role.IsAdmin && !strings.EqualFold(user.Role.Name, "administrator") {
					admin = " (administrator)"
				}
				fmt.Fprintf(out, "\nverification:  OK — authenticated as %s <%s>, role %s%s\n",
					user.Username, user.Email, orNone(user.Role.Name), admin)
			} else {
				fmt.Fprintf(out, "\nverification:  OK\n")
			}
			if v := client.MauticVersion(); v != "" {
				fmt.Fprintf(out, "mautic version: %s\n", v)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&verify, "verify", false, "call GET /users/self to prove the credentials actually work")
	return cmd
}

func identityOf(c *config.Credentials) string {
	if c.Method == config.AuthOAuth2 {
		if c.ClientID != "" {
			return "oauth2 client " + c.ClientID
		}
		return "oauth2 bearer token"
	}
	return c.Username
}

func newAuthLoginCommand() *cobra.Command {
	var (
		redirect     string
		noBrowser    bool
		clientCreds  bool
		saveProfile  string
		loginTimeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Obtain OAuth2 tokens and store them in a profile",
		Long: `Obtain OAuth2 tokens and store them in a profile.

Two grants are supported, because Mautic supports two:

  authorization_code (default)  opens a browser, needs a person, and returns a
                                refresh token valid for 14 days by default
  --client-credentials          no browser and no person, so it is the one that
                                works from CI — but not every instance enables it

The redirect URI must be registered in Mautic under Settings -> API Credentials
exactly as it is passed here, including the port and path.

Basic auth needs none of this: "mautic configure --username <user> --password
<key>" is enough, and is the simpler choice for scripts.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			profileName := config.ResolveProfileName(cfg, firstNonEmptyStr(saveProfile, flagProfile))

			baseURL, clientID, clientSecret := flagURL, flagClientID, flagClientSecret
			if prof, ok := cfg.Profiles[profileName]; ok {
				baseURL = firstNonEmptyStr(baseURL, prof.BaseURL)
				clientID = firstNonEmptyStr(clientID, prof.ClientID)
				if redirect == "" {
					redirect = prof.Redirect
				}
			}
			secrets, err := config.LoadSecrets()
			if err != nil {
				return err
			}
			if sec, ok := secrets.Profiles[profileName]; ok {
				clientSecret = firstNonEmptyStr(clientSecret, sec.ClientSecret)
			}
			baseURL = config.NormalizeBaseURL(baseURL)

			var missing []string
			if baseURL == "" {
				missing = append(missing, "instance URL (--url)")
			}
			if clientID == "" {
				missing = append(missing, "client ID (--client-id)")
			}
			if clientSecret == "" {
				missing = append(missing, "client secret (--client-secret)")
			}
			if len(missing) > 0 {
				return fmt.Errorf("cannot start the OAuth2 flow; missing:\n  - %s\n  create a client in Mautic under Settings -> API Credentials", joinLines(missing))
			}
			if redirect == "" {
				redirect = mautic.DefaultRedirectURI
			}

			var tok *tokenSet
			if clientCreds {
				t, err := mautic.ClientCredentials(ctx, baseURL, clientID, clientSecret)
				if err != nil {
					return err
				}
				tok = &tokenSet{t.AccessToken, t.RefreshToken, t.Expiry}
			} else {
				t, err := mautic.Login(ctx, mautic.LoginOptions{
					Config:      mautic.OAuth2Config(baseURL, clientID, clientSecret, redirect),
					OpenBrowser: !noBrowser,
					Timeout:     loginTimeout,
					Out:         cmd.ErrOrStderr(),
				})
				if err != nil {
					return err
				}
				tok = &tokenSet{t.AccessToken, t.RefreshToken, t.Expiry}
			}

			if profileName == "" {
				profileName = "default"
			}
			prof := cfg.Profiles[profileName]
			prof.BaseURL = baseURL
			prof.Auth = config.AuthOAuth2
			prof.ClientID = clientID
			prof.Redirect = redirect
			if cfg.Profiles == nil {
				cfg.Profiles = map[string]config.Profile{}
			}
			cfg.Profiles[profileName] = prof
			if cfg.DefaultProfile == "" {
				cfg.DefaultProfile = profileName
			}
			if err := cfg.Save(); err != nil {
				return err
			}

			sec := secrets.Profiles[profileName]
			sec.ClientSecret = clientSecret
			sec.AccessToken = tok.access
			if tok.refresh != "" {
				sec.RefreshToken = tok.refresh
			}
			sec.Expiry = tok.expiry
			secrets.Profiles[profileName] = sec
			if err := secrets.Save(); err != nil {
				return err
			}

			path, _ := config.SecretsPath()
			fmt.Fprintf(cmd.OutOrStdout(), "stored OAuth2 tokens for profile %q in %s\n", profileName, path)
			if tok.refresh == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "note: no refresh token was issued, so this access token cannot be renewed automatically\n")
			}
			return nil
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&redirect, "redirect-uri", "", "callback URL, registered verbatim in Mautic (default: "+mautic.DefaultRedirectURI+")")
	fl.BoolVar(&noBrowser, "no-browser", false, "print the authorize URL instead of opening a browser")
	fl.BoolVar(&clientCreds, "client-credentials", false, "use the client_credentials grant: no browser, no user, works from CI")
	fl.StringVar(&saveProfile, "save-profile", "", "profile to store the tokens in (default: the selected profile, or \"default\")")
	fl.DurationVar(&loginTimeout, "login-timeout", 5*time.Minute, "how long to wait for the browser callback")
	return cmd
}

type tokenSet struct {
	access  string
	refresh string
	expiry  time.Time
}

func newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Forget the stored secrets for a profile",
		Long: `Forget the stored secrets for a profile.

This clears the entry in credentials.json. The profile itself stays in
config.toml, so "mautic auth login" or "mautic configure" can refill it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			name := config.ResolveProfileName(cfg, flagProfile)
			if name == "" {
				return fmt.Errorf("no profile selected; pass --profile <name>")
			}
			secrets, err := config.LoadSecrets()
			if err != nil {
				return err
			}
			if _, ok := secrets.Profiles[name]; !ok {
				fmt.Fprintf(cmd.OutOrStdout(), "profile %q had no stored secrets\n", name)
				return nil
			}
			delete(secrets.Profiles, name)
			if err := secrets.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "cleared the stored secrets for profile %q\n", name)
			return nil
		},
	}
}

// errSilent reports failure through the exit code without printing a second
// error line, for commands that have already explained themselves in full.
type errSilent struct{}

func (errSilent) Error() string { return "" }

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func joinLines(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += "\n  - "
		}
		out += v
	}
	return out
}
