package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/output"
	"github.com/euxaristia/gitee-cli/internal/util"
)

func newPRCmd(app *App) *cobra.Command {
	prCmd := &cobra.Command{Use: "pr", Short: "Work with pull requests"}

	var repo, state string
	var page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				repo, _ = util.CurrentRepo()
			}
			if repo == "" {
				return fmt.Errorf("--repo is required or must be run inside a git repository")
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
		},
	}
	listCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	listCmd.Flags().StringVar(&state, "state", "open", "open|closed|all")
	listCmd.Flags().IntVar(&page, "page", 1, "Page number")
	listCmd.Flags().IntVar(&perPage, "per-page", 30, "Page size")

	viewCmd := &cobra.Command{
		Use:   "view [<number>]",
		Short: "View pull request",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				repo, _ = util.CurrentRepo()
			}
			if repo == "" {
				return fmt.Errorf("--repo is required or must be run inside a git repository")
			}
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, name, err := util.SplitRepo(repo)
			if err != nil {
				return err
			}

			var num int64
			if len(args) == 1 {
				n, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return err
				}
				num = n
			} else {
				branch, err := util.CurrentBranch()
				if err != nil {
					return fmt.Errorf("could not determine current branch: %w", err)
				}
				prs, err := app.Client.ListPRs(app.Ctx, owner, name, "open", owner+":"+branch, 1, 30)
				if err != nil {
					return err
				}
				if len(prs) == 0 {
					prs, err = app.Client.ListPRs(app.Ctx, owner, name, "open", branch, 1, 30)
					if err != nil {
						return err
					}
					if len(prs) == 0 {
						return fmt.Errorf("no open pull requests found for branch %s", branch)
					}
				}
				num = prs[0].Number
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
		},
	}
	viewCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")

	var title, body, head, base string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" || head == "" || base == "" {
				return fmt.Errorf("--title, --head and --base are required")
			}
			if repo == "" {
				repo, _ = util.CurrentRepo()
			}
			if repo == "" {
				return fmt.Errorf("--repo is required or must be run inside a git repository")
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
		},
	}
	createCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	createCmd.Flags().StringVar(&title, "title", "", "PR title")
	createCmd.Flags().StringVar(&body, "body", "", "PR body")
	createCmd.Flags().StringVar(&head, "head", "", "Head branch")
	createCmd.Flags().StringVar(&base, "base", "", "Base branch")

	var mergeTitle string
	mergeCmd := &cobra.Command{
		Use:   "merge <number>",
		Short: "Merge pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			if repo == "" {
				repo, _ = util.CurrentRepo()
			}
			if repo == "" {
				return fmt.Errorf("--repo is required or must be run inside a git repository")
			}
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, name, err := util.SplitRepo(repo)
			if err != nil {
				return err
			}
			if err := app.Client.MergePR(app.Ctx, owner, name, num, mergeTitle); err != nil {
				return err
			}
			fmt.Println("Merged")
			return nil
		},
	}
	mergeCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	mergeCmd.Flags().StringVar(&mergeTitle, "message", "", "Merge message")

	closeCmd := &cobra.Command{
		Use:   "close <number>",
		Short: "Close pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			if repo == "" {
				repo, _ = util.CurrentRepo()
			}
			if repo == "" {
				return fmt.Errorf("--repo is required or must be run inside a git repository")
			}
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, name, err := util.SplitRepo(repo)
			if err != nil {
				return err
			}
			pr, err := app.Client.ClosePR(app.Ctx, owner, name, num)
			if err != nil {
				return err
			}
			fmt.Printf("PR #%d -> %s\n", pr.Number, pr.State)
			return nil
		},
	}
	closeCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")

	commentCmd := &cobra.Command{
		Use:   "comment <number>",
		Short: "Comment on pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			if repo == "" {
				repo, _ = util.CurrentRepo()
			}
			if repo == "" {
				return fmt.Errorf("--repo is required or must be run inside a git repository")
			}
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, name, err := util.SplitRepo(repo)
			if err != nil {
				return err
			}
			if err := app.Client.CreatePRComment(app.Ctx, owner, name, num, body); err != nil {
				return err
			}
			fmt.Println("Comment added")
			return nil
		},
	}
	commentCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	commentCmd.Flags().StringVar(&body, "body", "", "Comment body")

	prCmd.AddCommand(listCmd, viewCmd, createCmd, mergeCmd, closeCmd, commentCmd)
	return prCmd
}
