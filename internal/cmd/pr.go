package cmd

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/output"
	"github.com/euxaristia/gitee-cli/internal/util"
)

func resolveRepo(repo *string) error {
	if *repo != "" {
		return nil
	}
	r, err := util.CurrentRepo()
	if err != nil || r == "" {
		return fmt.Errorf("--repo is required or must be run inside a git repository")
	}
	*repo = r
	return nil
}

func findPRForBranch(app *App, owner, name, branch string) (*api.PullRequest, error) {
	prs, err := app.Client.ListPRs(app.Ctx, owner, name, "open", owner+":"+branch, 1, 30)
	if err != nil {
		return nil, err
	}
	if len(prs) > 0 {
		return &prs[0], nil
	}

	prs, err = app.Client.ListPRs(app.Ctx, owner, name, "open", "", 1, 100)
	if err != nil {
		return nil, err
	}
	for _, pr := range prs {
		if pr.Head.Ref == branch || strings.HasSuffix(pr.Head.Label, ":"+branch) {
			return &pr, nil
		}
	}
	return nil, fmt.Errorf("no open pull requests found for branch %s", branch)
}

func resolvePRNumber(app *App, owner, name string, args []string) (int64, error) {
	if len(args) == 1 {
		return strconv.ParseInt(args[0], 10, 64)
	}
	branch, err := util.CurrentBranch()
	if err != nil {
		return 0, fmt.Errorf("could not determine current branch: %w", err)
	}
	pr, err := findPRForBranch(app, owner, name, branch)
	if err != nil {
		return 0, err
	}
	return pr.Number, nil
}

func newPRCmd(app *App) *Command {
	return &Command{
		Use:   "pr",
		Short: "Work with pull requests",
		run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("pr requires a subcommand")
			}
			if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
				return runHelp(app, c, append([]string{"pr"}, args[1:]...))
			}
			switch args[0] {
			case "list":
				return runPRList(app, c, args[1:])
			case "view":
				return runPRView(app, c, args[1:])
			case "create":
				return runPRCreate(app, c, args[1:])
			case "merge":
				return runPRMerge(app, c, args[1:])
			case "close":
				return runPRClose(app, c, args[1:])
			case "comment":
				return runPRComment(app, c, args[1:])
			default:
				return fmt.Errorf("unknown pr command %q", args[0])
			}
		},
	}
}

func runPRList(app *App, c *Command, args []string) error {
	pos, repo, state, page, perPage, err := parseRepoFlags("pr list", args, nil)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"pr", "list"})
		}
		return err
	}
	if err := exactArgs(pos, 0); err != nil {
		return err
	}
	if err := resolveRepo(&repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	prs, err := app.Client.ListPRs(app.Ctx, owner, name, state, "", page, perPage)
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(prs))
	for _, pr := range prs {
		rows = append(rows, []string{fmt.Sprintf("%d", pr.Number), pr.State, pr.Title, pr.User.Login})
	}
	return printAny(app.Cfg.Output, []string{"NUMBER", "STATE", "TITLE", "AUTHOR"}, rows, prs)
}

func runPRView(app *App, c *Command, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("pr view", args, nil)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"pr", "view"})
		}
		return err
	}
	if err := maxArgs(pos, 1); err != nil {
		return err
	}
	if err := resolveRepo(&repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}

	var num int64
	if len(pos) == 1 {
		n, err := strconv.ParseInt(pos[0], 10, 64)
		if err != nil {
			return err
		}
		num = n
	} else {
		branch, err := util.CurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}
		pr, err := findPRForBranch(app, owner, name, branch)
		if err != nil {
			return err
		}
		num = pr.Number
	}

	pr, err := app.Client.GetPR(app.Ctx, owner, name, num)
	if err != nil {
		return err
	}
	if app.Cfg.Output == string(output.FormatJSON) {
		return output.PrintJSON(pr)
	}
	fmt.Printf("#%d %s\n", pr.Number, pr.Title)
	fmt.Printf("State: %s\nAuthor: %s\nURL: %s\n", pr.State, pr.User.Login, pr.HTMLURL)
	return nil
}

func runPRCreate(app *App, c *Command, args []string) error {
	var title, body, head, base string
	pos, repo, _, _, _, err := parseRepoFlags("pr create", args, func(fs *flag.FlagSet) {
		fs.StringVar(&title, "title", "", "PR title")
		fs.StringVar(&body, "body", "", "PR body")
		fs.StringVar(&head, "head", "", "Head branch")
		fs.StringVar(&base, "base", "", "Base branch")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"pr", "create"})
		}
		return err
	}
	if err := exactArgs(pos, 0); err != nil {
		return err
	}
	if title == "" || head == "" || base == "" {
		return fmt.Errorf("--title, --head and --base are required")
	}
	if err := resolveRepo(&repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	pr, err := app.Client.CreatePR(app.Ctx, owner, name, title, head, base, body)
	if err != nil {
		return err
	}
	if app.Cfg.Output == string(output.FormatJSON) {
		return output.PrintJSON(pr)
	}
	fmt.Println(pr.HTMLURL)
	return nil
}

func runPRMerge(app *App, c *Command, args []string) error {
	var mergeTitle string
	pos, repo, _, _, _, err := parseRepoFlags("pr merge", args, func(fs *flag.FlagSet) {
		fs.StringVar(&mergeTitle, "message", "", "Merge message")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"pr", "merge"})
		}
		return err
	}
	if err := maxArgs(pos, 1); err != nil {
		return err
	}
	if err := resolveRepo(&repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	num, err := resolvePRNumber(app, owner, name, pos)
	if err != nil {
		return err
	}
	if err := app.Client.MergePR(app.Ctx, owner, name, num, mergeTitle); err != nil {
		return err
	}
	fmt.Println("Merged")
	return nil
}

func runPRClose(app *App, c *Command, args []string) error {
	pos, repo, _, _, _, err := parseRepoFlags("pr close", args, nil)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"pr", "close"})
		}
		return err
	}
	if err := maxArgs(pos, 1); err != nil {
		return err
	}
	if err := resolveRepo(&repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	num, err := resolvePRNumber(app, owner, name, pos)
	if err != nil {
		return err
	}
	pr, err := app.Client.ClosePR(app.Ctx, owner, name, num)
	if err != nil {
		return err
	}
	fmt.Printf("PR #%d -> %s\n", pr.Number, pr.State)
	return nil
}

func runPRComment(app *App, c *Command, args []string) error {
	var body string
	pos, repo, _, _, _, err := parseRepoFlags("pr comment", args, func(fs *flag.FlagSet) {
		fs.StringVar(&body, "body", "", "Comment body")
	})
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printHelp(c.OutOrStdout(), []string{"pr", "comment"})
		}
		return err
	}
	if err := maxArgs(pos, 1); err != nil {
		return err
	}
	if body == "" {
		return fmt.Errorf("--body is required")
	}
	if err := resolveRepo(&repo); err != nil {
		return err
	}
	if err := ensureToken(app); err != nil {
		return err
	}
	owner, name, err := util.SplitRepo(repo)
	if err != nil {
		return err
	}
	num, err := resolvePRNumber(app, owner, name, pos)
	if err != nil {
		return err
	}
	if err := app.Client.CreatePRComment(app.Ctx, owner, name, num, body); err != nil {
		return err
	}
	fmt.Println("Comment added")
	return nil
}
