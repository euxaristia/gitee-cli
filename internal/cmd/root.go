package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/auth"
	"github.com/euxaristia/gitee-cli/internal/config"
)

type App struct {
	Cfg         *config.Config
	Client      *api.Client
	ActiveToken string
	Ctx         context.Context
	GitRunner   gitRunner
}

func NewRootCmd() *Command {
	app := &App{Ctx: context.Background()}
	return &Command{
		Use:          "gt",
		Short:        "A full-featured CLI for Gitee",
		SilenceUsage: true,
		run: func(c *Command, args []string) error {
			rest, outputFormat, showVersion, isHelp, helpArgs, err := parseGlobal(args)
			if err != nil {
				return err
			}
			if showVersion {
				printVersionBanner(c.OutOrStdout())
				return nil
			}
			if isHelp {
				sub := newHelpCmd(app)
				sub.stdout = c.stdout
				sub.stderr = c.stderr
				sub.stdin = c.stdin
				sub.SetArgs(helpArgs)
				return sub.Execute()
			}
			if len(rest) == 0 {
				sub := newHelpCmd(app)
				sub.stdout = c.stdout
				sub.stderr = c.stderr
				sub.stdin = c.stdin
				sub.SetArgs(nil)
				return sub.Execute()
			}
			if err := initApp(app, outputFormat); err != nil {
				return err
			}
			return dispatchRoot(app, c, rest)
		},
	}
}

func parseGlobal(args []string) (rest []string, outputFormat string, showVersion bool, isHelp bool, helpArgs []string, err error) {
	var filtered []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			filtered = append(filtered, args[i+1:]...)
			i = len(args)
		case a == "-h" || a == "--help":
			isHelp = true
		case a == "-V" || a == "--version":
			showVersion = true
		case a == "-o" || a == "--output":
			if i+1 >= len(args) {
				return nil, "", false, false, nil, fmt.Errorf("flag needs an argument: %s", a)
			}
			i++
			outputFormat = args[i]
		case strings.HasPrefix(a, "-o="):
			outputFormat = strings.TrimPrefix(a, "-o=")
		case strings.HasPrefix(a, "--output="):
			outputFormat = strings.TrimPrefix(a, "--output=")
		default:
			filtered = append(filtered, a)
		}
	}
	if isHelp {
		return nil, outputFormat, showVersion, true, filtered, nil
	}
	return filtered, outputFormat, showVersion, false, nil, nil
}

func initApp(app *App, outputFormat string) error {
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
			storedToken = ""
		}
		if storedToken != "" {
			activeToken = storedToken
		} else {
			activeToken = cfg.Token
		}
	}
	app.Cfg = cfg
	app.ActiveToken = activeToken
	app.Client = api.New(cfg.APIBase, activeToken)
	return nil
}

func dispatchRoot(app *App, c *Command, args []string) error {
	name, rest := args[0], args[1:]
	var sub *Command
	switch name {
	case "auth":
		sub = newAuthCmd(app)
	case "git":
		sub = newGitCmd(app)
	case "commit":
		sub = newGitShortcutCmd(app, "commit")
	case "push":
		sub = newGitShortcutCmd(app, "push")
	case "pull":
		sub = newGitShortcutCmd(app, "pull")
	case "status":
		sub = newGitShortcutCmd(app, "status")
	case "repo":
		sub = newRepoCmd(app)
	case "issue":
		sub = newIssueCmd(app)
	case "pr":
		sub = newPRCmd(app)
	case "release":
		sub = newReleaseCmd(app)
	case "api":
		sub = newAPICmd(app)
	case "config":
		sub = newConfigCmd(app)
	case "completion":
		sub = newCompletionCmd()
	case "help":
		sub = newHelpCmd(app)
	case "version":
		sub = newVersionCmd()
	default:
		return fmt.Errorf("unknown command %q", name)
	}
	sub.stdout = c.stdout
	sub.stderr = c.stderr
	sub.stdin = c.stdin
	sub.SetArgs(rest)
	return sub.Execute()
}

func printRootHelp(w io.Writer) {
	_ = printHelp(w, nil)
}

func ensureToken(app *App) error {
	if env := os.Getenv("GITEE_TOKEN"); env != "" {
		app.ActiveToken = env
	}
	if app.ActiveToken == "" {
		return fmt.Errorf("no token configured: run `gt auth login` or set GITEE_TOKEN")
	}
	app.Client = api.New(app.Cfg.APIBase, app.ActiveToken)
	return nil
}
