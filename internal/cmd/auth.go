package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/auth"
	"github.com/euxaristia/gitee-cli/internal/config"
)

func newAuthCmd(app *App) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Gitee",
	}

	var token string
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login with a personal access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				readToken, err := auth.ReadTokenFromTTY()
				if err != nil {
					return err
				}
				token = readToken
			}
			if token == "" {
				return fmt.Errorf("empty token")
			}
			if err := auth.SaveToken(token); err != nil {
				return err
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			// Keep config token empty so secret lives in keychain.
			cfg.Token = ""
			if err := config.Save(cfg); err != nil {
				return err
			}
			app.ActiveToken = token
			app.Client = clientFrom(app, token)

			user, err := app.Client.CurrentUser(app.Ctx)
			if err == nil {
				cfg.User = user.Login
				config.Save(cfg)
			}

			fmt.Fprintln(os.Stdout, "Authenticated")
			return nil
		},
	}
	loginCmd.Flags().StringVar(&token, "token", "", "Gitee access token")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := auth.DeleteToken(); err != nil {
				return err
			}
			cfg.Token = ""
			cfg.User = ""
			if err := config.Save(cfg); err != nil {
				return err
			}
			app.ActiveToken = ""
			fmt.Fprintln(os.Stdout, "Logged out")
			return nil
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			host := app.Cfg.Host
			if host == "" {
				host = "https://gitee.com"
			}
			fmt.Fprintln(w, host)

			effectiveToken := app.ActiveToken
			if effectiveToken == "" {
				fmt.Fprintln(w, "  Not authenticated")
				fmt.Fprintf(w, "  Run `gt auth login` or set GITEE_TOKEN to authenticate.\n")
				return nil
			}

			source := tokenSource(app)
			user, err := app.Client.CurrentUser(app.Ctx)
			if err != nil {
				fmt.Fprintf(w, "  X Failed to verify token (%s)\n", source)
				return err
			}

			fmt.Fprintf(w, "  ✓ Logged in to %s account %s (%s)\n", host, user.Login, source)
			fmt.Fprintf(w, "    - Active account: true\n")
			fmt.Fprintf(w, "    - Git operations protocol: %s\n", app.Cfg.GitProtocol)
			fmt.Fprintf(w, "    - Token: %s\n", maskToken(effectiveToken))
			return nil
		},
	}

	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Print active token source",
		Run: func(cmd *cobra.Command, args []string) {
			if os.Getenv("GITEE_TOKEN") != "" {
				fmt.Fprintln(os.Stdout, "Using token from GITEE_TOKEN")
				return
			}
			if app.ActiveToken != "" {
				fmt.Fprintln(os.Stdout, "Using token from keychain")
				return
			}
			fmt.Fprintln(os.Stdout, "No token configured")
		},
	}

	gitCredentialCmd := &cobra.Command{
		Use:    "git-credential",
		Short:  "Helper for Git credentials",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			operation := args[0]
			if operation != "get" {
				return nil
			}

			scanner := bufio.NewScanner(os.Stdin)
			input := make(map[string]string)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					break
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					input[parts[0]] = parts[1]
				}
			}

			if host, ok := input["host"]; !ok || host != "gitee.com" {
				return nil
			}

			token := app.ActiveToken
			if token == "" {
				return nil
			}

			username := app.Cfg.User
			if username == "" {
				user, err := app.Client.CurrentUser(app.Ctx)
				if err == nil {
					username = user.Login
				} else {
					username = "oauth2"
				}
			}

			fmt.Fprintf(os.Stdout, "username=%s\n", username)
			fmt.Fprintf(os.Stdout, "password=%s\n", token)
			return nil
		},
	}

	setupGitCmd := &cobra.Command{
		Use:   "setup-git",
		Short: "Configure Git to use Gitee CLI as a credential helper",
		RunE: func(cmd *cobra.Command, args []string) error {
			exe, err := os.Executable()
			if err != nil {
				return err
			}

			gitCmd := exec.Command("git", "config", "--global", "credential.https://gitee.com.helper", "!"+exe+" auth git-credential")
			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("failed to configure git: %w", err)
			}

			fmt.Fprintln(os.Stdout, "Git configured to use Gitee CLI for credentials.")
			return nil
		},
	}

	authCmd.AddCommand(loginCmd, logoutCmd, statusCmd, tokenCmd, gitCredentialCmd, setupGitCmd)
	return authCmd
}

func clientFrom(app *App, token string) *api.Client {
	return api.New(app.Cfg.APIBase, token)
}

// tokenSource returns a human-readable label for where the active token comes from.
func tokenSource(app *App) string {
	if os.Getenv("GITEE_TOKEN") != "" {
		return "GITEE_TOKEN"
	}
	stored, err := auth.LoadToken()
	if err == nil && stored != "" {
		return "keyring"
	}
	if app.Cfg.Token != "" {
		return "config file"
	}
	return "unknown"
}

// maskToken returns a masked version of the token, showing only the first 4 characters.
func maskToken(token string) string {
	if len(token) <= 4 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-4)
}
