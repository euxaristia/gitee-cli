package cmd

import (
	"flag"
	"fmt"

	"github.com/euxaristia/gitee-cli/internal/output"
	"github.com/euxaristia/gitee-cli/internal/util"
)

func newIssueCmd(app *App) *Command {
	return &Command{
		Use:   "issue",
		Short: "Work with issues",
		run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("issue requires a subcommand")
			}
			switch args[0] {
			case "list":
				return runIssueList(app, args[1:])
			case "view":
				return runIssueView(app, args[1:])
			case "create":
				return runIssueCreate(app, args[1:])
			case "comment":
				return runIssueComment(app, args[1:])
			case "close":
				return runIssueClose(app, args[1:])
			case "reopen":
				return runIssueReopen(app, args[1:])
			case "status":
				return runIssueStatus(app)
			default:
				return fmt.Errorf("unknown issue command %q", args[0])
			}
		},
	}
}

func parseRepoFlags(name string, args []string, extra func(*flag.FlagSet)) (pos []string, repo, state string, page, perPage int, err error) {
	state = "open"
	page = 1
	perPage = 30
	pos, err = parseArgs(name, args, func(fs *flag.FlagSet) {
		fs.StringVar(&repo, "repo", "", "Repository owner/name")
		fs.StringVar(&state, "state", "open", "open|closed|all")
		fs.IntVar(&page, "page", 1, "Page number")
		fs.IntVar(&perPage, "per-page", 30, "Page size")
		if extra != nil {
			extra(fs)
		}
	})
	return pos, repo, state, page, perPage, err
}

func runIssueList(app *App, args []string) error {
	pos, repo, state, page, perPage, err := parseRepoFlags("issue list", args, nil)
	if err != nil {
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
	issues, err := app.Client.ListIssues(app.Ctx, owner, name, state, page, perPage)
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(issues))
	for _, i := range issues {
		rows = append(rows, []string{i.Number, i.State, i.Title, i.User.Login})
	}
	return printAny(app.Cfg.Output, []string{"NUMBER", "STATE", "TITLE", "AUTHOR"}, rows, issues)
}

func runIssueView(app *App, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("issue view", args, nil)
	if err != nil {
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
	issue, err := app.Client.GetIssue(app.Ctx, owner, name, pos[0])
	if err != nil {
		return err
	}
	if app.Cfg.Output == string(output.FormatJSON) {
		return output.PrintJSON(issue)
	}
	fmt.Printf("#%s %s\n", issue.Number, issue.Title)
	fmt.Printf("State: %s\nAuthor: %s\nURL: %s\n", issue.State, issue.User.Login, issue.HTMLURL)
	return nil
}

func runIssueCreate(app *App, args []string) error {
	var title, body string
	pos, repo, _, _, _, err := parseRepoFlags("issue create", args, func(fs *flag.FlagSet) {
		fs.StringVar(&title, "title", "", "Issue title")
		fs.StringVar(&body, "body", "", "Issue body")
	})
	if err != nil {
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
	if title == "" {
		return fmt.Errorf("--title is required")
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	issue, err := app.Client.CreateIssue(app.Ctx, owner, name, title, body)
	if err != nil {
		return err
	}
	if app.Cfg.Output == string(output.FormatJSON) {
		return output.PrintJSON(issue)
	}
	fmt.Println(issue.HTMLURL)
	return nil
}

func runIssueComment(app *App, args []string) error {
	var body string
	pos, repo, _, _, _, err := parseRepoFlags("issue comment", args, func(fs *flag.FlagSet) {
		fs.StringVar(&body, "body", "", "Comment body")
	})
	if err != nil {
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
	if body == "" {
		return fmt.Errorf("--body is required")
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	if err := app.Client.CreateIssueComment(app.Ctx, owner, name, pos[0], body); err != nil {
		return err
	}
	fmt.Println("Comment added")
	return nil
}

func runIssueClose(app *App, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("issue close", args, nil)
	if err != nil {
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	if err := requireFlag("repo", repo); err != nil {
		return err
	}
	return changeIssueState(app, repo, pos[0], "closed")
}

func runIssueReopen(app *App, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("issue reopen", args, nil)
	if err != nil {
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	if err := requireFlag("repo", repo); err != nil {
		return err
	}
	return changeIssueState(app, repo, pos[0], "open")
}

func runIssueStatus(app *App) error {
	if err := ensureToken(app); err != nil {
		return err
	}

	fmt.Println("Issues assigned to you")
	assigned, err := app.Client.ListAllIssues(app.Ctx, "assigned", "open", 1, 10)
	if err != nil {
		return err
	}
	if len(assigned) == 0 {
		fmt.Println("  None")
	} else {
		for _, i := range assigned {
			fmt.Printf("  #%s  %s\n", i.Number, i.Title)
		}
	}

	fmt.Println("\nIssues created by you")
	created, err := app.Client.ListAllIssues(app.Ctx, "created", "open", 1, 10)
	if err != nil {
		return err
	}
	if len(created) == 0 {
		fmt.Println("  None")
	} else {
		for _, i := range created {
			fmt.Printf("  #%s  %s\n", i.Number, i.Title)
		}
	}
	return nil
}

func changeIssueState(app *App, repo, number, state string) error {
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	issue, err := app.Client.UpdateIssueState(app.Ctx, owner, name, number, state)
	if err != nil {
		return err
	}
	fmt.Printf("Issue #%s -> %s\n", issue.Number, issue.State)
	return nil
}
