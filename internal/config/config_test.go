package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeBaseURL(t *testing.T) {
	// The two ideamans Mautic tools disagree about whether the configured URL
	// includes /api, so both spellings have to land on the same instance root.
	cases := map[string]string{
		"https://mautic.example.com":      "https://mautic.example.com",
		"https://mautic.example.com/":     "https://mautic.example.com",
		"https://mautic.example.com/api":  "https://mautic.example.com",
		"https://mautic.example.com/api/": "https://mautic.example.com",
		"  https://mautic.example.com  ":  "https://mautic.example.com",
		"":                                "",
	}
	for in, want := range cases {
		if got := NormalizeBaseURL(in); got != want {
			t.Errorf("NormalizeBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

// isolate points the config directory at a temporary path and clears every
// environment variable the resolver reads, so a developer's real credentials
// cannot make a test pass.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(envConfigDirVar, dir)
	for _, names := range [][]string{
		envBaseURL, envUsername, envPassword, envClientID,
		envClientSecret, envAccessToken, envRefreshToken, envRedirect, envProfile,
	} {
		for _, n := range names {
			t.Setenv(n, "")
		}
	}
	return dir
}

func TestResolveFromEnvironment(t *testing.T) {
	isolate(t)
	// The alias spellings my-mautic-assistant already uses.
	t.Setenv("MAUTIC_API_ENDPOINT", "https://mautic.example.com/api")
	t.Setenv("MAUTIC_API_USER", "api-user")
	t.Setenv("MAUTIC_API_KEY", "s3cret")

	creds, err := Resolve(Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if creds.Source != "env" {
		t.Errorf("Source = %q, want env", creds.Source)
	}
	if creds.BaseURL != "https://mautic.example.com" {
		t.Errorf("BaseURL = %q", creds.BaseURL)
	}
	if creds.APIURL() != "https://mautic.example.com/api" {
		t.Errorf("APIURL = %q", creds.APIURL())
	}
	if creds.Method != AuthBasic {
		t.Errorf("Method = %q, want basic", creds.Method)
	}
}

func TestEnvironmentBeatsProfile(t *testing.T) {
	isolate(t)
	cfg := &Config{
		DefaultProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {BaseURL: "https://prod.example.com", Auth: AuthBasic, Username: "prod-user"},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	secrets := &Secrets{Profiles: map[string]Secret{"prod": {Password: "prod-pass"}}}
	if err := secrets.Save(); err != nil {
		t.Fatalf("Save secrets: %v", err)
	}

	// With no environment set, the profile wins.
	creds, err := Resolve(Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if creds.Source != "profile:prod" || creds.BaseURL != "https://prod.example.com" {
		t.Fatalf("expected the prod profile, got %+v", creds)
	}

	// A complete environment set beats it. This is what makes CI work on a
	// machine that also has profiles.
	t.Setenv("MAUTIC_BASE_URL", "https://ci.example.com")
	t.Setenv("MAUTIC_API_USER", "ci-user")
	t.Setenv("MAUTIC_API_KEY", "ci-pass")
	creds, err = Resolve(Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if creds.Source != "env" || creds.BaseURL != "https://ci.example.com" || creds.Username != "ci-user" {
		t.Fatalf("environment did not win: %+v", creds)
	}
}

func TestFlagsBeatEverything(t *testing.T) {
	isolate(t)
	t.Setenv("MAUTIC_BASE_URL", "https://env.example.com")
	t.Setenv("MAUTIC_API_USER", "env-user")
	t.Setenv("MAUTIC_API_KEY", "env-pass")

	creds, err := Resolve(Overrides{BaseURL: "https://flag.example.com/api", Username: "flag-user"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if creds.BaseURL != "https://flag.example.com" {
		t.Errorf("BaseURL = %q, want the flag value normalised", creds.BaseURL)
	}
	if creds.Username != "flag-user" {
		t.Errorf("Username = %q, want flag-user", creds.Username)
	}
	if creds.Password != "env-pass" {
		t.Errorf("Password = %q, want the environment value to remain", creds.Password)
	}
}

func TestResolveWithNothingConfigured(t *testing.T) {
	isolate(t)
	_, err := Resolve(Overrides{})
	if err == nil {
		t.Fatal("expected an error when nothing is configured")
	}
	// The message has to name the fix, not just the problem.
	for _, want := range []string{"MAUTIC_BASE_URL", "mautic configure"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q:\n%v", want, err)
		}
	}
}

func TestMissingValuesAreListedTogether(t *testing.T) {
	isolate(t)
	// A base URL alone is not a usable credential set, and the error should
	// name both remaining values rather than one per round trip.
	_, err := Resolve(Overrides{BaseURL: "https://mautic.example.com", Auth: AuthBasic})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "MAUTIC_API_USER") || !strings.Contains(err.Error(), "MAUTIC_API_KEY") {
		t.Errorf("both missing values should be listed at once:\n%v", err)
	}
}

func TestOAuth2ProfileRoundTrip(t *testing.T) {
	isolate(t)
	cfg := &Config{
		DefaultProfile: "oauth",
		Profiles: map[string]Profile{
			"oauth": {BaseURL: "https://mautic.example.com", Auth: AuthOAuth2, ClientID: "1_abc"},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	secrets := &Secrets{Profiles: map[string]Secret{
		"oauth": {ClientSecret: "shh", AccessToken: "at", RefreshToken: "rt"},
	}}
	if err := secrets.Save(); err != nil {
		t.Fatalf("Save secrets: %v", err)
	}

	creds, err := Resolve(Overrides{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if creds.Method != AuthOAuth2 || creds.AccessToken != "at" || creds.ClientID != "1_abc" {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
	if strings.Contains(creds.Describe(), "shh") || strings.Contains(creds.Describe(), "at") {
		t.Errorf("Describe must never leak a secret: %q", creds.Describe())
	}
}

func TestSaveTokensLeavesOtherProfilesAlone(t *testing.T) {
	isolate(t)
	secrets := &Secrets{Profiles: map[string]Secret{
		"a": {Password: "a-pass"},
		"b": {RefreshToken: "old"},
	}}
	if err := secrets.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := SaveTokens("b", "new-access", "new-refresh", creds0()); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
	reloaded, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets: %v", err)
	}
	if reloaded.Profiles["a"].Password != "a-pass" {
		t.Error("SaveTokens clobbered another profile")
	}
	if reloaded.Profiles["b"].RefreshToken != "new-refresh" {
		t.Errorf("refresh token not rotated: %+v", reloaded.Profiles["b"])
	}
}

// SaveTokens with an empty profile means the credentials came from the
// environment; writing a file there would surprise a CI runner.
func TestSaveTokensIsANoOpForEnvironmentCredentials(t *testing.T) {
	dir := isolate(t)
	if err := SaveTokens("", "a", "b", creds0()); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, credsFileName)); !os.IsNotExist(err) {
		t.Error("SaveTokens wrote credentials.json for environment credentials")
	}
}

func TestSecretFilePermissions(t *testing.T) {
	dir := isolate(t)
	secrets := &Secrets{Profiles: map[string]Secret{"a": {Password: "p"}}}
	if err := secrets.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, credsFileName))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("credentials.json mode = %o, want 600", perm)
	}
}

func TestMissingFilesAreEmptyNotErrors(t *testing.T) {
	isolate(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load on a missing file should succeed: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected no profiles, got %v", cfg.ProfileNames())
	}
	secrets, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets on a missing file should succeed: %v", err)
	}
	if len(secrets.Profiles) != 0 {
		t.Error("expected no secrets")
	}
}

// creds0 is a zero expiry, spelled out so the test reads clearly.
func creds0() time.Time { return time.Time{} }

// encoding/json's omitempty does nothing for a struct, so a plain time.Time
// expiry would write "0001-01-01T00:00:00Z" into every basic-auth profile —
// a real-looking expiry that already passed.
func TestBasicAuthProfileHasNoExpiryField(t *testing.T) {
	isolate(t)
	secrets := &Secrets{Profiles: map[string]Secret{"default": {Password: "p"}}}
	if err := secrets.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	path, _ := SecretsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(data), "expiry") {
		t.Errorf("credentials.json should carry no expiry for basic auth:\n%s", data)
	}
}

// An OAuth2 profile must still round-trip its expiry.
func TestOAuth2ExpiryRoundTrips(t *testing.T) {
	isolate(t)
	want := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	if err := SaveTokens("oauth", "at", "rt", want); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
	secrets, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets: %v", err)
	}
	got := TimeValue(secrets.Profiles["oauth"].Expiry)
	if !got.Equal(want) {
		t.Errorf("expiry = %v, want %v", got, want)
	}
}

// A credentials.json written before Expiry became optional carries a zero
// timestamp. Loading must drop it, or `configure` writes it straight back.
func TestZeroExpiryIsHealedOnLoad(t *testing.T) {
	dir := isolate(t)
	legacy := `{"profiles":{"default":{"password":"p","expiry":"0001-01-01T00:00:00Z"}}}`
	if err := os.WriteFile(filepath.Join(dir, credsFileName), []byte(legacy), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	secrets, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets: %v", err)
	}
	if secrets.Profiles["default"].Expiry != nil {
		t.Error("a zero expiry should be dropped on load")
	}
	if err := secrets.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, credsFileName))
	if strings.Contains(string(data), "expiry") {
		t.Errorf("re-saving should not reintroduce expiry:\n%s", data)
	}
	if !strings.Contains(string(data), `"password": "p"`) {
		t.Errorf("the password must survive the heal:\n%s", data)
	}
}
