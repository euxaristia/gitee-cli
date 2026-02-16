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
			effectiveToken := app.ActiveToken
			if effectiveToken == "" {
				fmt.Fprintln(os.Stdout, "Not authenticated")
				return nil
			}
			user, err := app.Client.CurrentUser(app.Ctx)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Authenticated as %s (%s)\n", user.Login, user.Name)
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
