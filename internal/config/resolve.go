package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Environment variable names, newest/canonical first. Aliases are accepted
// quietly so that a shell already set up for my-mautic-assistant
// (MAUTIC_API_ENDPOINT / MAUTIC_API_USER / MAUTIC_API_KEY) or for
// mautic-common-oauth2 (MAUTIC_BASE_URL / MAUTIC_CLIENT_ID / ...) works with no
// changes. Error messages always name the first entry.
var (
	envBaseURL  = []string{"MAUTIC_BASE_URL", "MAUTIC_API_ENDPOINT", "MAUTIC_URL"}
	envUsername = []string{"MAUTIC_API_USER", "MAUTIC_USERNAME", "MAUTIC_USER"}
	envPassword = []string{"MAUTIC_API_KEY", "MAUTIC_PASSWORD", "MAUTIC_API_PASSWORD"}
	// Mautic's own API Credentials screen labels these "Public Key" and
	// "Secret Key", so that is what people paste into their .env. Accepting
	// those spellings turns "why doesn't it see my client id" into a non-event.
	envClientID     = []string{"MAUTIC_CLIENT_ID", "MAUTIC_OAUTH_CLIENT_ID", "MAUTIC_OAUTH2_PUBLIC_KEY", "MAUTIC_PUBLIC_KEY"}
	envClientSecret = []string{"MAUTIC_CLIENT_SECRET", "MAUTIC_OAUTH_CLIENT_SECRET", "MAUTIC_OAUTH2_SECRET_KEY", "MAUTIC_SECRET_KEY"}
	envAccessToken  = []string{"MAUTIC_ACCESS_TOKEN"}
	envRefreshToken = []string{"MAUTIC_REFRESH_TOKEN"}
	envRedirect     = []string{"MAUTIC_OAUTH2_REDIRECT_URI", "MAUTIC_REDIRECT_URI"}
	envProfile      = []string{"MAUTIC_PROFILE"}
)

// Overrides carries values that came from command line flags. Empty means
// "not specified", so they never blank out a configured value.
type Overrides struct {
	Profile      string
	BaseURL      string
	Auth         string
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
	AccessToken  string
}

// Credentials is one resolved identity, ready for the API client.
//
// Source records where it came from ("env" or "profile:<name>"). An agent that
// can say "using profile client-a" is an agent that stops before writing to the
// wrong Mautic instance, so every write path prints it.
type Credentials struct {
	Source  string
	Profile string // empty when Source is "env"

	BaseURL string // instance root, no trailing /api
	Method  string // AuthBasic or AuthOAuth2

	Username string
	Password string

	ClientID     string
	ClientSecret string
	Redirect     string
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// APIURL is the REST API root for this instance.
func (c *Credentials) APIURL() string { return c.BaseURL + "/api" }

// Describe renders a one-line, secret-free summary for `auth status` and for
// --dry-run output.
func (c *Credentials) Describe() string {
	who := c.Username
	if c.Method == AuthOAuth2 {
		who = "oauth2 client " + c.ClientID
	}
	return fmt.Sprintf("%s as %s (%s, source: %s)", c.BaseURL, who, c.Method, c.Source)
}

func env(names []string) string {
	for _, n := range names {
		if v := strings.TrimSpace(os.Getenv(n)); v != "" {
			return v
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Environment accessors for the OAuth2 login flow, which resolves its inputs
// before a profile exists and so cannot go through Resolve. Without these,
// `auth login` would be the one command that ignores the environment.
func EnvBaseURL() string      { return NormalizeBaseURL(env(envBaseURL)) }
func EnvClientID() string     { return env(envClientID) }
func EnvClientSecret() string { return env(envClientSecret) }
func EnvRedirect() string     { return env(envRedirect) }

// ResolveProfileName reports which profile would be used, without requiring it
// to exist. Precedence: --profile > MAUTIC_PROFILE > default_profile.
func ResolveProfileName(cfg *Config, flag string) string {
	return firstNonEmpty(flag, env(envProfile), cfg.DefaultProfile)
}

// Resolve builds the effective credentials.
//
// Order:
//  1. The environment, when it carries a complete set. It wins over every
//     profile so that CI and one-off runs need no config file at all.
//  2. The selected profile from config.toml + credentials.json.
//  3. An error that lists everything missing and names the command that fixes it.
//
// Flags override whichever of the two won.
func Resolve(o Overrides) (*Credentials, error) {
	if c := fromEnv(o); c != nil {
		applyOverrides(c, o)
		return c, c.validate()
	}

	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	name := ResolveProfileName(cfg, o.Profile)
	if name == "" {
		return nil, noProfileError(cfg)
	}
	prof, ok := cfg.Profiles[name]
	if !ok {
		path, _ := FilePath()
		return nil, fmt.Errorf("profile %q is not configured in %s\n  configured profiles: %s\n  add it with: mautic configure --profile %s --url https://mautic.example.com --username <user> --password <key>",
			name, path, describeNames(cfg.ProfileNames()), name)
	}

	secrets, err := LoadSecrets()
	if err != nil {
		return nil, err
	}
	sec := secrets.Profiles[name]

	c := &Credentials{
		Source:       "profile:" + name,
		Profile:      name,
		BaseURL:      NormalizeBaseURL(prof.BaseURL),
		Method:       firstNonEmpty(prof.Auth, AuthBasic),
		Username:     prof.Username,
		Password:     sec.Password,
		ClientID:     prof.ClientID,
		ClientSecret: sec.ClientSecret,
		Redirect:     prof.Redirect,
		AccessToken:  sec.AccessToken,
		RefreshToken: sec.RefreshToken,
		Expiry:       TimeValue(sec.Expiry),
	}
	applyOverrides(c, o)
	return c, c.validate()
}

// fromEnv returns credentials when the environment carries a complete set,
// otherwise nil so that Resolve falls through to the profiles.
func fromEnv(o Overrides) *Credentials {
	base := NormalizeBaseURL(firstNonEmpty(o.BaseURL, env(envBaseURL)))
	if base == "" {
		return nil
	}

	user, pass := env(envUsername), env(envPassword)
	token := env(envAccessToken)
	clientID, clientSecret := env(envClientID), env(envClientSecret)
	refresh := env(envRefreshToken)

	c := &Credentials{
		Source:       "env",
		BaseURL:      base,
		Username:     user,
		Password:     pass,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  token,
		RefreshToken: refresh,
		Redirect:     env(envRedirect),
	}

	// An explicit --auth decides; otherwise pick by what is actually present.
	switch {
	case o.Auth == AuthOAuth2, token != "", refresh != "" && clientID != "":
		c.Method = AuthOAuth2
	case user != "" && pass != "":
		c.Method = AuthBasic
	default:
		return nil
	}
	return c
}

func applyOverrides(c *Credentials, o Overrides) {
	if o.BaseURL != "" {
		c.BaseURL = NormalizeBaseURL(o.BaseURL)
	}
	if o.Auth != "" {
		c.Method = o.Auth
	}
	c.Username = firstNonEmpty(o.Username, c.Username)
	c.Password = firstNonEmpty(o.Password, c.Password)
	c.ClientID = firstNonEmpty(o.ClientID, c.ClientID)
	c.ClientSecret = firstNonEmpty(o.ClientSecret, c.ClientSecret)
	c.AccessToken = firstNonEmpty(o.AccessToken, c.AccessToken)
	if c.Method == "" {
		c.Method = AuthBasic
	}
}

// validate lists everything missing at once. One value per round trip is what
// makes an agent give up.
func (c *Credentials) validate() error {
	var missing []string
	if c.BaseURL == "" {
		missing = append(missing, "instance URL (--url or MAUTIC_BASE_URL)")
	}
	switch c.Method {
	case AuthBasic:
		if c.Username == "" {
			missing = append(missing, "username (--username or MAUTIC_API_USER)")
		}
		if c.Password == "" {
			missing = append(missing, "password (--password or MAUTIC_API_KEY)")
		}
	case AuthOAuth2:
		if c.AccessToken == "" && c.RefreshToken == "" {
			missing = append(missing, "access token (run `mautic auth login`, or set MAUTIC_ACCESS_TOKEN)")
		}
		if c.AccessToken == "" && c.RefreshToken != "" && (c.ClientID == "" || c.ClientSecret == "") {
			missing = append(missing, "OAuth2 client (--client-id/--client-secret or MAUTIC_CLIENT_ID/MAUTIC_CLIENT_SECRET)")
		}
	default:
		return fmt.Errorf("unknown auth method %q; use %q or %q", c.Method, AuthBasic, AuthOAuth2)
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("credentials are incomplete (%s):\n  - %s\n  fix with: mautic configure --url https://mautic.example.com --username <user> --password <key>",
		c.Source, strings.Join(missing, "\n  - "))
}

func noProfileError(cfg *Config) error {
	path, _ := FilePath()
	if len(cfg.Profiles) > 0 {
		return fmt.Errorf("no profile selected and no default is set\n  configured profiles: %s\n  pick one with: mautic profiles use <name>",
			describeNames(cfg.ProfileNames()))
	}
	return fmt.Errorf("no credentials configured (%s does not exist)\n  set MAUTIC_BASE_URL, MAUTIC_API_USER and MAUTIC_API_KEY,\n  or run: mautic configure --url https://mautic.example.com --username <user> --password <key>",
		path)
}

func describeNames(names []string) string {
	if len(names) == 0 {
		return "(none)"
	}
	return strings.Join(names, ", ")
}
