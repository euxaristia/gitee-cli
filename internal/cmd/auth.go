package cmd

import (
	"fmt"
	"os"

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

	authCmd.AddCommand(loginCmd, logoutCmd, statusCmd, tokenCmd)
	return authCmd
}

func clientFrom(app *App, token string) *api.Client {
	return api.New(app.Cfg.APIBase, token)
}
