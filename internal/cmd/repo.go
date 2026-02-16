package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/output"
	"github.com/euxaristia/gitee-cli/internal/util"
)

func newRepoCmd(app *App) *cobra.Command {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Work with repositories",
	}

	var org, visibility string
	var page, perPage int
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureToken(app); err != nil {
				return err
			}
			repos, err := app.Client.ListRepos(app.Ctx, org, visibility, page, perPage)
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(repos))
			for _, r := range repos {
				rows = append(rows, []string{r.FullName, r.DefaultBr, r.HTMLURL})
			}
			return printAny(app.Cfg.Output, []string{"NAME", "DEFAULT", "URL"}, rows, repos)
		},
	}
	listCmd.Flags().StringVar(&org, "org", "", "Organization name")
	listCmd.Flags().StringVar(&visibility, "visibility", "all", "all|public|private")
	listCmd.Flags().IntVar(&page, "page", 1, "Page number")
	listCmd.Flags().IntVar(&perPage, "per-page", 30, "Page size")

	viewCmd := &cobra.Command{
		Use:   "view <owner/repo>",
		Short: "View repository details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureToken(app); err != nil {
				return err
			}
			owner, repo, err := util.SplitRepo(args[0])
			if err != nil {
				return err
			}
			r, err := app.Client.GetRepo(app.Ctx, owner, repo)
			if err != nil {
				return err
			}
			rows := [][]string{{r.FullName, r.DefaultBr, fmt.Sprintf("%t", r.Private), r.HTMLURL}}
			return printAny(app.Cfg.Output, []string{"NAME", "DEFAULT", "PRIVATE", "URL"}, rows, r)
		},
	}

	var name, desc string
	var private bool
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Println(r.HTMLURL)
			return nil
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "Repository name")
	createCmd.Flags().StringVar(&desc, "description", "", "Repository description")
	createCmd.Flags().StringVar(&org, "org", "", "Organization name")
	createCmd.Flags().BoolVar(&private, "private", false, "Create private repository")

	var dest string
	cloneCmd := &cobra.Command{
		Use:   "clone <owner/repo>",
		Short: "Clone repository via git",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := util.SplitRepo(args[0])
			if err != nil {
				return err
			}
			url := fmt.Sprintf("https://gitee.com/%s/%s.git", owner, repo)
			gitArgs := []string{"clone", url}
			if dest != "" {
				gitArgs = append(gitArgs, dest)
			}
			c := exec.Command("git", gitArgs...)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()
			return c.Run()
		},
	}
	cloneCmd.Flags().StringVar(&dest, "dest", "", "Destination directory")

	repoCmd.AddCommand(listCmd, viewCmd, createCmd, cloneCmd)
	return repoCmd
}
