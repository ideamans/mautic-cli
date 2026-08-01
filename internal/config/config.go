// Package config reads and writes the CLI's settings and credentials.
//
// Everything lives under ~/.config/mautic-cli/ (honouring XDG_CONFIG_HOME),
// split into a non-secret file and a secret one:
//
//	config.toml       ... profile list, default profile, base URL, auth method
//	credentials.json  ... passwords, client secrets and OAuth2 tokens (0600)
//
// The split is deliberate: config.toml can be pasted into a support thread,
// and the file rewritten by `mautic configure` is not the file holding tokens.
//
// Values resolve in the order flag > environment variable > config file. The
// environment beats every profile, which is what lets one code path serve both
// a laptop with five profiles and a CI runner with none.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	// dirName is the service, not the binary — see the house standard.
	appDirName      = "mautic-cli"
	configFileName  = "config.toml"
	credsFileName   = "credentials.json"
	envConfigDirVar = "MAUTIC_CLI_HOME"
)

// Auth methods. Mautic supports both; which one an instance accepts is decided
// by its administrator under Configuration -> API Settings.
const (
	AuthBasic  = "basic"
	AuthOAuth2 = "oauth2"
)

// Profile is one named Mautic instance. Secrets are NOT stored here; they live
// in credentials.json under the same profile name.
type Profile struct {
	BaseURL  string `toml:"base_url"`
	Auth     string `toml:"auth"`
	Username string `toml:"username,omitempty"`
	ClientID string `toml:"client_id,omitempty"`
	// Redirect is the OAuth2 callback URL registered in Mautic. It must match
	// the portal registration exactly, so the port cannot be random.
	Redirect string `toml:"redirect_uri,omitempty"`
}

// Config is the whole of config.toml.
type Config struct {
	DefaultProfile string             `toml:"default_profile,omitempty"`
	Profiles       map[string]Profile `toml:"profiles,omitempty"`
}

// Secret holds one profile's secret material.
type Secret struct {
	Password     string    `json:"password,omitempty"`
	ClientSecret string    `json:"client_secret,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

// Secrets is the whole of credentials.json.
type Secrets struct {
	Profiles map[string]Secret `json:"profiles"`
}

// Dir returns the configuration directory.
// Precedence: MAUTIC_CLI_HOME > $XDG_CONFIG_HOME/mautic-cli > ~/.config/mautic-cli.
func Dir() (string, error) {
	if d := os.Getenv(envConfigDirVar); d != "" {
		return d, nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine the home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appDirName), nil
}

// FilePath returns the path of config.toml.
func FilePath() (string, error) { return join(configFileName) }

// SecretsPath returns the path of credentials.json.
func SecretsPath() (string, error) { return join(credsFileName) }

func join(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func ensureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("could not create the config directory (%s): %w", dir, err)
	}
	return dir, nil
}

// Load reads config.toml. A missing file is an empty config, not an error —
// every command must work with nothing configured.
func Load() (*Config, error) {
	path, err := FilePath()
	if err != nil {
		return nil, err
	}
	c := &Config{Profiles: map[string]Profile{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}
	if err := toml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	return c, nil
}

// Save writes config.toml with the directory at 0700 and the file at 0600.
func (c *Config) Save() error {
	if _, err := ensureDir(); err != nil {
		return err
	}
	path, err := FilePath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("could not write %s: %w", path, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("could not encode %s: %w", path, err)
	}
	return nil
}

// ProfileNames returns the configured profile names, sorted so that output is
// deterministic.
func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LoadSecrets reads credentials.json. A missing file is empty, not an error.
func LoadSecrets() (*Secrets, error) {
	path, err := SecretsPath()
	if err != nil {
		return nil, err
	}
	s := &Secrets{Profiles: map[string]Secret{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}
	if s.Profiles == nil {
		s.Profiles = map[string]Secret{}
	}
	return s, nil
}

// Save writes credentials.json at 0600.
func (s *Secrets) Save() error {
	if _, err := ensureDir(); err != nil {
		return err
	}
	path, err := SecretsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode credentials: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("could not write %s: %w", path, err)
	}
	return nil
}

// SaveTokens persists an OAuth2 token set for one profile, leaving the rest of
// the file alone. The API client calls this after a silent refresh so the
// rotated refresh token is not lost.
func SaveTokens(profile, accessToken, refreshToken string, expiry time.Time) error {
	if profile == "" {
		// Credentials came from the environment; there is nothing to update,
		// and writing a file would surprise a CI runner.
		return nil
	}
	s, err := LoadSecrets()
	if err != nil {
		return err
	}
	sec := s.Profiles[profile]
	sec.AccessToken = accessToken
	if refreshToken != "" {
		sec.RefreshToken = refreshToken
	}
	sec.Expiry = expiry
	s.Profiles[profile] = sec
	return s.Save()
}

// NormalizeBaseURL turns anything a user is likely to paste into the instance
// root, with no trailing slash and no trailing /api.
//
// This matters more than it looks: the two ideamans Mautic tools disagree.
// my-mautic-assistant's MAUTIC_API_ENDPOINT is "https://host/api" while
// mautic-common-oauth2's MAUTIC_BASE_URL is "https://host". Accepting both and
// normalising here is what stops every request 404ing on /api/api/contacts.
func NormalizeBaseURL(raw string) string {
	u := strings.TrimSpace(raw)
	u = strings.TrimRight(u, "/")
	u = strings.TrimSuffix(u, "/api")
	return strings.TrimRight(u, "/")
}
