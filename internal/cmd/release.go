package cmd

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/euxaristia/gitee-cli/internal/util"
)

func newReleaseCmd(app *App) *Command {
	return &Command{
		Use:   "release",
		Short: "Manage releases",
		run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("release requires a subcommand")
			}
			if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
				return runHelp(app, c, append([]string{"release"}, args[1:]...))
			}
			switch args[0] {
			case "list":
				return runReleaseList(app, c, args[1:])
			case "view":
				return runReleaseView(app, c, args[1:])
			case "create":
				return runReleaseCreate(app, c, args[1:])
			case "delete":
				return runReleaseDelete(app, c, args[1:])
			default:
				return fmt.Errorf("unknown release command %q", args[0])
			}
		},
	}
}

func runReleaseList(app *App, c *Command, args []string) error {
	pos, repo, _, page, perPage, err := parseRepoFlags("release list", args, nil)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"release", "list"})
		}
		return err
	}
	if err := exactArgs(pos, 0); err != nil {
		return err
	}
	if err := requireFlag("repo", repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	releases, err := app.Client.ListReleases(app.Ctx, owner, name, page, perPage)
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(releases))
	for _, r := range releases {
		rows = append(rows, []string{r.TagName, r.Name, r.HTMLURL})
	}
	return printAny(app.Cfg.Output, []string{"TAG", "NAME", "URL"}, rows, releases)
}

func runReleaseView(app *App, c *Command, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("release view", args, nil)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"release", "view"})
		}
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	if err := requireFlag("repo", repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	rel, err := app.Client.GetReleaseByTag(app.Ctx, owner, name, pos[0])
	if err != nil {
		return err
	}
	rows := [][]string{{rel.TagName, rel.Name, rel.HTMLURL}}
	return printAny(app.Cfg.Output, []string{"TAG", "NAME", "URL"}, rows, rel)
}

func runReleaseCreate(app *App, c *Command, args []string) error {
	var tag, name, body, target string
	var draft bool
	pos, repo, _, _, _, err := parseRepoFlags("release create", args, func(fs *flag.FlagSet) {
		fs.StringVar(&tag, "tag", "", "Tag name")
		fs.StringVar(&name, "name", "", "Release name")
		fs.StringVar(&body, "body", "", "Release notes")
		fs.StringVar(&target, "target", "", "Target commitish")
		fs.BoolVar(&draft, "draft", false, "Create as draft")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"release", "create"})
		}
		return err
	}
	if err := exactArgs(pos, 0); err != nil {
		return err
	}
	if err := requireFlag("repo", repo); err != nil {
		return err
	}
	if tag == "" {
		return fmt.Errorf("--tag is required")
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, repoName, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	rel, err := app.Client.CreateRelease(app.Ctx, owner, repoName, tag, name, body, target, draft)
	if err != nil {
		return err
	}
	fmt.Println(rel.HTMLURL)
	return nil
}

func runReleaseDelete(app *App, c *Command, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("release delete", args, nil)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"release", "delete"})
		}
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	id, err := strconv.ParseInt(pos[0], 10, 64)
	if err != nil {
		return err
	}
	if err := requireFlag("repo", repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, repoName, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	if err := app.Client.DeleteRelease(app.Ctx, owner, repoName, id); err != nil {
		return err
	}
	fmt.Println("Deleted")
	return nil
}
