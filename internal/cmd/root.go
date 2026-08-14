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
			rest, outputFormat, showVersion, help, err := parseGlobal(args)
			if err != nil {
				return err
			}
			if showVersion {
				printVersionBanner(c.OutOrStdout())
				return nil
			}
			if help || len(rest) == 0 {
				printRootHelp(c.OutOrStdout())
				return nil
			}
			if err := initApp(app, outputFormat); err != nil {
				return err
			}
			return dispatchRoot(app, c, rest)
		},
	}
}

func parseGlobal(args []string) (rest []string, outputFormat string, showVersion, help bool, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			return args[i+1:], outputFormat, showVersion, help, nil
		case a == "-h" || a == "--help":
			return nil, outputFormat, showVersion, true, nil
		case a == "-V" || a == "--version":
			showVersion = true
		case a == "-o" || a == "--output":
			if i+1 >= len(args) {
				return nil, "", false, false, fmt.Errorf("flag needs an argument: %s", a)
			}
			i++
			outputFormat = args[i]
		case strings.HasPrefix(a, "-o="):
			outputFormat = strings.TrimPrefix(a, "-o=")
		case strings.HasPrefix(a, "--output="):
			outputFormat = strings.TrimPrefix(a, "--output=")
		case strings.HasPrefix(a, "-"):
			return args[i:], outputFormat, showVersion, help, nil
		default:
			return args[i:], outputFormat, showVersion, help, nil
		}
	}
	return nil, outputFormat, showVersion, help, nil
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
	fmt.Fprintln(w, "A full-featured CLI for Gitee")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  gt [flags]")
	fmt.Fprintln(w, "  gt [command]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available Commands:")
	fmt.Fprintln(w, "  api         Make a raw API request to Gitee v5")
	fmt.Fprintln(w, "  auth        Authenticate with Gitee")
	fmt.Fprintln(w, "  commit      Run `git commit`")
	fmt.Fprintln(w, "  completion  Generate shell completion scripts")
	fmt.Fprintln(w, "  config      Manage gitee CLI config")
	fmt.Fprintln(w, "  git         Run git operations with retry for transient network failures")
	fmt.Fprintln(w, "  issue       Work with issues")
	fmt.Fprintln(w, "  pr          Work with pull requests")
	fmt.Fprintln(w, "  pull        Run `git pull`")
	fmt.Fprintln(w, "  push        Run `git push`")
	fmt.Fprintln(w, "  release     Manage releases")
	fmt.Fprintln(w, "  repo        Work with repositories")
	fmt.Fprintln(w, "  status      Run `git status`")
	fmt.Fprintln(w, "  version     Print version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -h, --help            help")
	fmt.Fprintln(w, "  -o, --output string   Output format: table|json")
	fmt.Fprintln(w, "  -V, --version         Print version information")
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
