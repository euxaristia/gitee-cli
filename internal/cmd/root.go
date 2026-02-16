package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/auth"
	"github.com/euxaristia/gitee-cli/internal/config"
)

type App struct {
	Cfg         *config.Config
	Client      *api.Client
	ActiveToken string
	Ctx         context.Context
}

func NewRootCmd() *cobra.Command {
	app := &App{Ctx: context.Background()}
	var outputFormat string

	root := &cobra.Command{
		Use:   "gitee",
		Short: "A full-featured CLI for Gitee",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if outputFormat != "" {
				cfg.Output = outputFormat
			}
			activeToken := ""
			if env := os.Getenv("GITEE_TOKEN"); env != "" {
				activeToken = env
			} else {
				storedToken, err := auth.LoadToken()
				if err != nil {
					return err
				}
				if storedToken != "" {
					activeToken = storedToken
				} else {
					// Legacy fallback for previously stored plaintext tokens.
					activeToken = cfg.Token
				}
			}
			app.Cfg = cfg
			app.ActiveToken = activeToken
			app.Client = api.New(cfg.APIBase, activeToken)
			return nil
		},
	}

	root.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: table|json")

	root.AddCommand(
		newAuthCmd(app),
		newRepoCmd(app),
		newIssueCmd(app),
		newPRCmd(app),
		newReleaseCmd(app),
		newAPICmd(app),
		newConfigCmd(app),
		newCompletionCmd(root),
		newVersionCmd(),
	)

	root.SetHelpCommand(&cobra.Command{Hidden: true})
	root.SilenceUsage = true
	return root
}

func ensureToken(app *App) error {
	if env := os.Getenv("GITEE_TOKEN"); env != "" {
		app.ActiveToken = env
	}
	if app.ActiveToken == "" {
		return fmt.Errorf("no token configured: run `gitee auth login` or set GITEE_TOKEN")
	}
	app.Client = api.New(app.Cfg.APIBase, app.ActiveToken)
	return nil
}
