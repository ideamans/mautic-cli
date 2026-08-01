// Package cmd assembles the mautic command tree.
//
// Commands parse flags and format output; they never build HTTP requests by
// hand. Everything goes through internal/mautic, which is what keeps the
// wrapped commands and the raw `mautic api` escape hatch on one code path.
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/ideamans/mautic-cli/internal/config"
	"github.com/ideamans/mautic-cli/internal/mautic"
)

// PluginVersion is the single version source. plugins/mautic-cli's plugin.json
// must match it (enforced by plugin_test.go) and so must the git tag (enforced
// by the release workflow).
const PluginVersion = "0.1.0"

// version is what --version reports. goreleaser overwrites it at build time
// with -X github.com/ideamans/mautic-cli/cmd.version.
var version = PluginVersion

// Build metadata, injected by goreleaser.
var (
	commit = ""
	date   = ""
)

// Global flag values.
var (
	flagProfile      string
	flagURL          string
	flagAuth         string
	flagUsername     string
	flagPassword     string
	flagClientID     string
	flagClientSecret string
	flagAccessToken  string

	flagFormat  string
	flagFields  string
	flagDryRun  bool
	flagYes     bool
	flagTimeout time.Duration
	flagRetries int
	flagVerbose bool
)

// Root returns the fully assembled command tree without executing it.
//
// The catalog generator needs exactly this, and retrofitting it later is the
// most-forgotten step in the house standard — so it exists from the first
// commit.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "mautic",
		Short: "Command line client for the Mautic REST API",
		Long: `mautic wraps the Mautic REST API so that scripts and AI agents can drive a
Mautic instance without hand-assembling HTTP requests.

It is non-interactive: nothing prompts, destructive commands require --yes, and
anything that writes accepts --dry-run. stdout carries the result and nothing
else; progress and warnings go to stderr.

Run "mautic llm" for the full reference, including which command answers which
question, the field aliases, and the failure modes that produce confidently
wrong answers.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       versionString(),
	}
	root.SetVersionTemplate("{{.Version}}\n")

	pf := root.PersistentFlags()
	pf.StringVar(&flagProfile, "profile", "", "named profile to use (env: MAUTIC_PROFILE)")
	pf.StringVar(&flagURL, "url", "", "Mautic instance URL, with or without a trailing /api (env: MAUTIC_BASE_URL, MAUTIC_API_ENDPOINT)")
	pf.StringVar(&flagAuth, "auth", "", "authentication method: basic or oauth2 (default: whatever the profile or environment provides)")
	pf.StringVar(&flagUsername, "username", "", "Mautic username for basic auth (env: MAUTIC_API_USER)")
	pf.StringVar(&flagPassword, "password", "", "Mautic password for basic auth (env: MAUTIC_API_KEY)")
	pf.StringVar(&flagClientID, "client-id", "", "OAuth2 client ID (env: MAUTIC_CLIENT_ID)")
	pf.StringVar(&flagClientSecret, "client-secret", "", "OAuth2 client secret (env: MAUTIC_CLIENT_SECRET)")
	pf.StringVar(&flagAccessToken, "access-token", "", "OAuth2 access token to use as-is (env: MAUTIC_ACCESS_TOKEN)")

	pf.StringVar(&flagFormat, "format", "table", "output format: table, json, jsonl or csv")
	pf.StringVar(&flagFields, "fields", "", "comma-separated dotted paths selecting columns, e.g. id,fields.all.email")
	pf.BoolVar(&flagDryRun, "dry-run", false, "print the request that would be sent and send nothing")
	pf.BoolVar(&flagYes, "yes", false, "skip the confirmation gate on destructive commands")
	pf.DurationVar(&flagTimeout, "timeout", 60*time.Second, "timeout for a single HTTP request")
	pf.IntVar(&flagRetries, "retries", 3, "extra attempts for HTTP 429 and 5xx responses")
	pf.BoolVar(&flagVerbose, "verbose", false, "log each request to stderr")

	root.AddCommand(
		newConfigureCommand(),
		newProfilesCommand(),
		newAuthCommand(),
		newAPICommand(),
	)
	// Everything that follows the standard CRUD shape comes from the resource
	// table; the rest are endpoints whose shape the table cannot express.
	root.AddCommand(resourceCommands()...)
	root.AddCommand(
		newFieldsCommand(),
		newFilesCommand(),
		newThemesCommand(),
		newStatsCommand(),
		newDataCommand(),
	)
	return root
}

func versionString() string {
	s := version
	if commit != "" {
		s += " (" + commit
		if date != "" {
			s += ", " + date
		}
		s += ")"
	}
	return s
}

// Version returns the version this binary reports.
func Version() string { return version }

// overrides collects the credential flags.
func overrides() config.Overrides {
	return config.Overrides{
		Profile:      flagProfile,
		BaseURL:      flagURL,
		Auth:         flagAuth,
		Username:     flagUsername,
		Password:     flagPassword,
		ClientID:     flagClientID,
		ClientSecret: flagClientSecret,
		AccessToken:  flagAccessToken,
	}
}

// resolve returns the effective credentials, or an error naming everything that
// is missing and the command that fixes it.
func resolve() (*config.Credentials, error) {
	return config.Resolve(overrides())
}

// newClient resolves credentials and builds an API client from them.
func newClient(ctx context.Context) (*mautic.Client, *config.Credentials, error) {
	creds, err := resolve()
	if err != nil {
		return nil, nil, err
	}
	auth, err := authenticator(ctx, creds)
	if err != nil {
		return nil, nil, err
	}
	var trace io.Writer
	if flagVerbose {
		trace = os.Stderr
	}
	client := mautic.New(mautic.Options{
		BaseURL: creds.BaseURL,
		Auth:    auth,
		Timeout: flagTimeout,
		Retries: flagRetries,
		DryRun:  flagDryRun,
		Trace:   trace,
	})
	return client, creds, nil
}

// authenticator turns resolved credentials into something that can sign a
// request, wiring OAuth2 token rotation back into credentials.json.
func authenticator(ctx context.Context, creds *config.Credentials) (mautic.Authenticator, error) {
	switch creds.Method {
	case config.AuthBasic:
		return &mautic.BasicAuth{Username: creds.Username, Password: creds.Password}, nil
	case config.AuthOAuth2:
		cfg := mautic.OAuth2Config(creds.BaseURL, creds.ClientID, creds.ClientSecret, creds.Redirect)
		tok := &oauth2.Token{
			AccessToken:  creds.AccessToken,
			RefreshToken: creds.RefreshToken,
			TokenType:    "Bearer",
			Expiry:       creds.Expiry,
		}
		profile := creds.Profile
		persist := func(access, refresh string, expiry time.Time) error {
			return config.SaveTokens(profile, access, refresh, expiry)
		}
		return mautic.NewOAuth2Auth(ctx, cfg, tok, persist), nil
	default:
		return nil, fmt.Errorf("unknown auth method %q; use basic or oauth2", creds.Method)
	}
}

// confirmed gates a destructive action. There are no prompts anywhere in this
// CLI, so the only way through is --yes.
func confirmed(cmd *cobra.Command, action string) error {
	if flagYes || flagDryRun {
		return nil
	}
	return fmt.Errorf("%s is destructive and was not performed\n  re-run with --yes to confirm, or with --dry-run to see the request first", action)
}

// noteTarget prints which instance and identity a write is about to touch.
// "I would delete 400 contacts" and "I would delete 400 contacts on the
// production instance" are different sentences.
func noteTarget(cmd *cobra.Command, creds *config.Credentials) {
	fmt.Fprintf(cmd.ErrOrStderr(), "target: %s\n", creds.Describe())
}
