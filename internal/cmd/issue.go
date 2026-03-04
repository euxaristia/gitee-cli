package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/output"
	"github.com/euxaristia/gitee-cli/internal/util"
)

func newIssueCmd(app *App) *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Work with issues",
	}

	var repo, state string
	var page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
	listCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	listCmd.Flags().StringVar(&state, "state", "open", "open|closed|all")
	listCmd.Flags().IntVar(&page, "page", 1, "Page number")
	listCmd.Flags().IntVar(&perPage, "per-page", 30, "Page size")
	_ = listCmd.MarkFlagRequired("repo")

	viewCmd := &cobra.Command{
		Use:   "view <number>",
		Short: "View issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, name, err := util.SplitRepo(repo)
			if err != nil {
				return err
			}
			issue, err := app.Client.GetIssue(app.Ctx, owner, name, args[0])
			if err != nil {
				return err
			}
			if app.Cfg.Output == string(output.FormatJSON) {
				return output.PrintJSON(issue)
			}
			fmt.Printf("#%s %s\n", issue.Number, issue.Title)
			fmt.Printf("State: %s\nAuthor: %s\nURL: %s\n", issue.State, issue.User.Login, issue.HTMLURL)
			return nil
		},
	}
	viewCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	_ = viewCmd.MarkFlagRequired("repo")

	var title, body string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create issue",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
	createCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	createCmd.Flags().StringVar(&title, "title", "", "Issue title")
	createCmd.Flags().StringVar(&body, "body", "", "Issue body")
	_ = createCmd.MarkFlagRequired("repo")

	commentCmd := &cobra.Command{
		Use:   "comment <number>",
		Short: "Comment on issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if err := app.Client.CreateIssueComment(app.Ctx, owner, name, args[0], body); err != nil {
				return err
			}
			fmt.Println("Comment added")
			return nil
		},
	}
	commentCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	commentCmd.Flags().StringVar(&body, "body", "", "Comment body")
	_ = commentCmd.MarkFlagRequired("repo")

	closeCmd := &cobra.Command{
		Use:   "close <number>",
		Short: "Close issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return changeIssueState(app, repo, args[0], "closed")
		},
	}
	closeCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	_ = closeCmd.MarkFlagRequired("repo")

	reopenCmd := &cobra.Command{
		Use:   "reopen <number>",
		Short: "Reopen issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return changeIssueState(app, repo, args[0], "open")
		},
	}
	reopenCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	_ = reopenCmd.MarkFlagRequired("repo")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show a summary of issues relevant to you",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}

	issueCmd.AddCommand(listCmd, viewCmd, createCmd, commentCmd, closeCmd, reopenCmd, statusCmd)
	return issueCmd
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
