package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/util"
)

func newReleaseCmd(app *App) *cobra.Command {
	releaseCmd := &cobra.Command{Use: "release", Short: "Manage releases"}

	var repo string
	var page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List releases",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
	listCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	listCmd.Flags().IntVar(&page, "page", 1, "Page number")
	listCmd.Flags().IntVar(&perPage, "per-page", 30, "Page size")
	_ = listCmd.MarkFlagRequired("repo")

	viewCmd := &cobra.Command{
		Use:   "view <tag>",
		Short: "View release by tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, name, err := util.SplitRepo(repo)
			if err != nil {
				return err
			}
			rel, err := app.Client.GetReleaseByTag(app.Ctx, owner, name, args[0])
			if err != nil {
				return err
			}
			rows := [][]string{{rel.TagName, rel.Name, rel.HTMLURL}}
			return printAny(app.Cfg.Output, []string{"TAG", "NAME", "URL"}, rows, rel)
		},
	}
	viewCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	_ = viewCmd.MarkFlagRequired("repo")

	var tag, name, body, target string
	var draft bool
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a release",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
	createCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	createCmd.Flags().StringVar(&tag, "tag", "", "Tag name")
	createCmd.Flags().StringVar(&name, "name", "", "Release name")
	createCmd.Flags().StringVar(&body, "body", "", "Release notes")
	createCmd.Flags().StringVar(&target, "target", "", "Target commitish")
	createCmd.Flags().BoolVar(&draft, "draft", false, "Create as draft")
	_ = createCmd.MarkFlagRequired("repo")

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete release by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
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
		},
	}
	deleteCmd.Flags().StringVar(&repo, "repo", "", "Repository owner/name")
	_ = deleteCmd.MarkFlagRequired("repo")

	releaseCmd.AddCommand(listCmd, viewCmd, createCmd, deleteCmd)
	return releaseCmd
}
