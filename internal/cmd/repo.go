package cmd

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/euxaristia/gitee-cli/internal/output"
	"github.com/euxaristia/gitee-cli/internal/util"
)

func newRepoCmd(app *App) *Command {
	return &Command{
		Use:   "repo",
		Short: "Work with repositories",
		run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("repo requires a subcommand")
			}
			if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
				return runHelp(app, c, append([]string{"repo"}, args[1:]...))
			}
			switch args[0] {
			case "list":
				return runRepoList(app, c, args[1:])
			case "view":
				return runRepoView(app, c, args[1:])
			case "create":
				return runRepoCreate(app, c, args[1:])
			case "clone":
				return runRepoClone(app, c, args[1:])
			default:
				return fmt.Errorf("unknown repo command %q", args[0])
			}
		},
	}
}

func runRepoList(app *App, c *Command, args []string) error {
	var org, visibility string
	var page, perPage int
	pos, err := parseArgs("repo list", args, func(fs *flag.FlagSet) {
		fs.StringVar(&org, "org", "", "Organization name")
		fs.StringVar(&visibility, "visibility", "all", "all|public|private")
		fs.IntVar(&page, "page", 1, "Page number")
		fs.IntVar(&perPage, "per-page", 30, "Page size")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"repo", "list"})
		}
		return err
	}
	if err := exactArgs(pos, 0); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	repos, err := app.Client.ListRepos(app.Ctx, org, visibility, page, perPage)
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(repos))
	for _, r := range repos {
		visibility := "public"
		if r.Private {
			visibility = "private"
		}
		rows = append(rows, []string{r.FullName, visibility, r.DefaultBr, strings.TrimSuffix(r.HTMLURL, ".git")})
	}
	return printAny(app.Cfg.Output, []string{"NAME", "VISIBILITY", "DEFAULT", "URL"}, rows, repos)
}

func runRepoView(app *App, c *Command, args []string) error {
	pos, err := parseArgs("repo view", args, func(fs *flag.FlagSet) {})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"repo", "view"})
		}
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, repo, err := util.SplitRepo(pos[0])
	if err != nil {
		return err
	}
	r, err := app.Client.GetRepo(app.Ctx, owner, repo)
	if err != nil {
		return err
	}
	rows := [][]string{{r.FullName, r.DefaultBr, fmt.Sprintf("%t", r.Private), strings.TrimSuffix(r.HTMLURL, ".git")}}
	return printAny(app.Cfg.Output, []string{"NAME", "DEFAULT", "PRIVATE", "URL"}, rows, r)
}

func runRepoCreate(app *App, c *Command, args []string) error {
	var name, desc, org string
	var private bool
	pos, err := parseArgs("repo create", args, func(fs *flag.FlagSet) {
		fs.StringVar(&name, "name", "", "Repository name")
		fs.StringVar(&desc, "description", "", "Repository description")
		fs.StringVar(&org, "org", "", "Organization name")
		fs.BoolVar(&private, "private", false, "Create private repository")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"repo", "create"})
		}
		return err
	}
	if err := exactArgs(pos, 0); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("--name is required")
	}
	r, err := app.Client.CreateRepo(app.Ctx, name, desc, org, private)
	if err != nil {
		return err
	}
	if app.Cfg.Output == string(output.FormatJSON) {
		return output.PrintJSON(r)
	}
	fmt.Println(strings.TrimSuffix(r.HTMLURL, ".git"))
	return nil
}

func runRepoClone(app *App, c *Command, args []string) error {
	var dest string
	var depth int
	var recursive, useSSH bool
	pos, err := parseArgs("repo clone", args, func(fs *flag.FlagSet) {
		fs.StringVar(&dest, "dest", "", "Destination directory")
		fs.IntVar(&depth, "depth", 0, "Create a shallow clone")
		fs.BoolVar(&recursive, "recursive", false, "Clone submodules")
		fs.BoolVar(&useSSH, "ssh", false, "Use SSH protocol for cloning")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"repo", "clone"})
		}
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	owner, repo, err := util.SplitRepo(pos[0])
	if err != nil {
		return err
	}

	protocol := app.Cfg.GitProtocol
	if useSSH {
		protocol = "ssh"
	}

	var url string
	if protocol == "ssh" {
		url = fmt.Sprintf("git@gitee.com:%s/%s.git", owner, repo)
	} else {
		url = fmt.Sprintf("https://gitee.com/%s/%s.git", owner, repo)
	}

	gitArgs := []string{url}
	if depth > 0 {
		gitArgs = append(gitArgs, "--depth", fmt.Sprintf("%d", depth))
	}
	if recursive {
		gitArgs = append(gitArgs, "--recursive")
	}
	if dest != "" {
		gitArgs = append(gitArgs, dest)
	}
	return runGitWithRetry(app, c, "clone", gitArgs)
}
