package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ideamans/mautic-cli/internal/config"
)

func newConfigureCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Add or update a named profile",
		Long: `Add or update a named profile.

A profile is one Mautic instance plus the credentials for it. The non-secret
half goes to config.toml and the secrets to credentials.json, both under
~/.config/mautic-cli/ at 0600.

Basic auth is the quickest way to a working setup:

  mautic configure --url https://mautic.example.com \
    --username api-user --password s3cret

For OAuth2, register the client here and then run "mautic auth login":

  mautic configure --url https://mautic.example.com --auth oauth2 \
    --client-id 1_abc --client-secret def
  mautic auth login

Nothing here is required to use the CLI: setting MAUTIC_BASE_URL,
MAUTIC_API_USER and MAUTIC_API_KEY works with no config file at all, and the
environment beats every profile.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			profileName := firstNonEmptyStr(name, flagProfile, cfg.DefaultProfile, "default")

			prof := cfg.Profiles[profileName]
			if flagURL != "" {
				prof.BaseURL = config.NormalizeBaseURL(flagURL)
			}
			if flagAuth != "" {
				prof.Auth = flagAuth
			}
			if prof.Auth == "" {
				prof.Auth = config.AuthBasic
			}
			if flagUsername != "" {
				prof.Username = flagUsername
			}
			if flagClientID != "" {
				prof.ClientID = flagClientID
			}

			if prof.BaseURL == "" {
				return fmt.Errorf("--url is required the first time a profile is configured")
			}
			switch prof.Auth {
			case config.AuthBasic:
				if prof.Username == "" {
					return fmt.Errorf("--username is required for basic auth")
				}
			case config.AuthOAuth2:
				if prof.ClientID == "" {
					return fmt.Errorf("--client-id is required for oauth2; create a client in Mautic under Settings -> API Credentials")
				}
			default:
				return fmt.Errorf("--auth must be basic or oauth2, got %q", prof.Auth)
			}

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

			secrets, err := config.LoadSecrets()
			if err != nil {
				return err
			}
			sec := secrets.Profiles[profileName]
			if flagPassword != "" {
				sec.Password = flagPassword
			}
			if flagClientSecret != "" {
				sec.ClientSecret = flagClientSecret
			}
			if flagAccessToken != "" {
				sec.AccessToken = flagAccessToken
			}
			if secrets.Profiles == nil {
				secrets.Profiles = map[string]config.Secret{}
			}
			secrets.Profiles[profileName] = sec
			if err := secrets.Save(); err != nil {
				return err
			}

			path, _ := config.FilePath()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "saved profile %q to %s\n", profileName, path)
			fmt.Fprintf(out, "  instance: %s\n", prof.BaseURL)
			fmt.Fprintf(out, "  auth:     %s\n", prof.Auth)
			if prof.Auth == config.AuthBasic && sec.Password == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "no password is stored for this profile — pass --password, or set MAUTIC_API_KEY\n")
			}
			if prof.Auth == config.AuthOAuth2 && sec.AccessToken == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "no token is stored yet — run `mautic auth login`\n")
			}
			fmt.Fprintf(out, "\nverify it with: mautic auth status --verify\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile name to write (default: --profile, then the current default, then \"default\")")
	return cmd
}

func newProfilesCommand() *cobra.Command {
	group := &cobra.Command{
		Use:   "profiles",
		Short: "List, switch and remove named profiles",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List the configured profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			secrets, err := config.LoadSecrets()
			if err != nil {
				return err
			}
			names := cfg.ProfileNames()
			if len(names) == 0 {
				path, _ := config.FilePath()
				fmt.Fprintf(cmd.ErrOrStderr(), "no profiles configured in %s\n  add one with: mautic configure --url https://mautic.example.com --username <user> --password <key>\n", path)
				return nil
			}
			sort.Strings(names)
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDEFAULT\tINSTANCE\tAUTH\tIDENTITY\tSECRET")
			for _, n := range names {
				p := cfg.Profiles[n]
				def := ""
				if n == cfg.DefaultProfile {
					def = "*"
				}
				identity := p.Username
				if p.Auth == config.AuthOAuth2 {
					identity = p.ClientID
				}
				sec := secrets.Profiles[n]
				stored := "missing"
				switch {
				case p.Auth == config.AuthOAuth2 && sec.AccessToken != "":
					stored = "token"
				case p.Auth == config.AuthBasic && sec.Password != "":
					stored = "password"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", n, def, p.BaseURL, p.Auth, orNone(identity), stored)
			}
			return w.Flush()
		},
	}

	use := &cobra.Command{
		Use:   "use <name>",
		Short: "Set the default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q is not configured\n  configured profiles: %v", args[0], cfg.ProfileNames())
			}
			cfg.DefaultProfile = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "default profile is now %q\n", args[0])
			return nil
		},
	}

	remove := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a profile and its stored secrets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmed(cmd, "removing profile "+args[0]); err != nil {
				return err
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q is not configured", args[0])
			}
			delete(cfg.Profiles, args[0])
			if cfg.DefaultProfile == args[0] {
				cfg.DefaultProfile = ""
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			secrets, err := config.LoadSecrets()
			if err != nil {
				return err
			}
			delete(secrets.Profiles, args[0])
			if err := secrets.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed profile %q\n", args[0])
			if cfg.DefaultProfile == "" && len(cfg.Profiles) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "there is no default profile now — set one with: mautic profiles use <name>\n")
			}
			return nil
		},
	}

	group.AddCommand(list, use, remove)
	return group
}
