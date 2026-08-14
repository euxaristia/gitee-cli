package cmd

import (
	"fmt"
	"io"
	"strings"
)

type helpTopic struct {
	Use      string
	Short    string
	Commands []helpCommand
	Flags    []helpFlag
}

type helpCommand struct {
	Name        string
	Description string
}

type helpFlag struct {
	Short       string
	Long        string
	Type        string
	Description string
}

func newHelpCmd(app *App) *Command {
	return &Command{
		Use:   "help [command]",
		Short: "Help about any command",
		run: func(c *Command, args []string) error {
			return runHelp(app, c, args)
		},
	}
}

func runHelp(app *App, c *Command, args []string) error {
	w := c.OutOrStdout()
	return printHelp(w, args)
}

func printHelp(w io.Writer, args []string) error {
	var cleanArgs []string
	for _, a := range args {
		if a != "--help" && a != "-h" {
			cleanArgs = append(cleanArgs, a)
		}
	}

	key := strings.Join(cleanArgs, " ")
	key = strings.TrimSpace(key)

	topic, ok := helpTopics[key]
	if !ok {
		if len(cleanArgs) > 0 {
			return fmt.Errorf("unknown command %q", key)
		}
		return fmt.Errorf("no help available")
	}

	fmt.Fprintln(w, topic.Short)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	for _, u := range strings.Split(topic.Use, "\n") {
		fmt.Fprintf(w, "  %s\n", u)
	}

	if len(topic.Commands) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Available Commands:")
		cmdWidth := 12
		for _, cmd := range topic.Commands {
			if len(cmd.Name) > cmdWidth-2 {
				cmdWidth = len(cmd.Name) + 2
			}
		}
		for _, cmd := range topic.Commands {
			padding := strings.Repeat(" ", cmdWidth-len(cmd.Name))
			fmt.Fprintf(w, "  %s%s%s\n", cmd.Name, padding, cmd.Description)
		}
	}

	if len(topic.Flags) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Flags:")
		for _, f := range topic.Flags {
			var fStr string
			if f.Short != "" && f.Long != "" {
				fStr = fmt.Sprintf("%s, %s", f.Short, f.Long)
			} else if f.Long != "" {
				fStr = fmt.Sprintf("    %s", f.Long)
			} else if f.Short != "" {
				fStr = fmt.Sprintf("%s", f.Short)
			}
			if f.Type != "" {
				fStr += " " + f.Type
			}
			padding := ""
			if len(fStr) < 22 {
				padding = strings.Repeat(" ", 22-len(fStr))
			} else {
				padding = "  "
			}
			fmt.Fprintf(w, "  %s%s%s\n", fStr, padding, f.Description)
		}
	}

	return nil
}

var helpTopics = map[string]helpTopic{
	"": {
		Short: "A full-featured CLI for Gitee",
		Use:   "gt [flags]\ngt [command]",
		Commands: []helpCommand{
			{Name: "api", Description: "Make a raw API request to Gitee v5"},
			{Name: "auth", Description: "Authenticate with Gitee"},
			{Name: "commit", Description: "Run `git commit`"},
			{Name: "completion", Description: "Generate shell completion scripts"},
			{Name: "config", Description: "Manage gitee CLI config"},
			{Name: "git", Description: "Run git operations with retry for transient network failures"},
			{Name: "help", Description: "Help about any command"},
			{Name: "issue", Description: "Work with issues"},
			{Name: "pr", Description: "Work with pull requests"},
			{Name: "pull", Description: "Run `git pull`"},
			{Name: "push", Description: "Run `git push`"},
			{Name: "release", Description: "Manage releases"},
			{Name: "repo", Description: "Work with repositories"},
			{Name: "status", Description: "Run `git status`"},
			{Name: "version", Description: "Print version"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help"},
			{Short: "-o", Long: "--output", Type: "string", Description: "Output format: table|json"},
			{Short: "-V", Long: "--version", Description: "Print version information"},
		},
	},
	"api": {
		Short: "Make a raw API request to Gitee v5",
		Use:   "gt api <endpoint> [flags]",
		Flags: []helpFlag{
			{Short: "-F", Long: "--field", Type: "string", Description: "Add key=value request field"},
			{Short: "-H", Long: "--header", Type: "string", Description: "Add request header key:value"},
			{Short: "-h", Long: "--help", Description: "help for api"},
			{Short: "-X", Long: "--method", Type: "string", Description: "HTTP method (default \"GET\")"},
		},
	},
	"auth": {
		Short: "Authenticate with Gitee",
		Use:   "gt auth <command>",
		Commands: []helpCommand{
			{Name: "git-credential", Description: "Git credential helper"},
			{Name: "login", Description: "Authenticate with Gitee"},
			{Name: "logout", Description: "Log out of Gitee"},
			{Name: "setup-git", Description: "Configure git to use Gitee CLI for credentials"},
			{Name: "status", Description: "View authentication status"},
			{Name: "token", Description: "Print current access token"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for auth"},
		},
	},
	"auth git-credential": {
		Short: "Git credential helper",
		Use:   "gt auth git-credential",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for git-credential"},
		},
	},
	"auth login": {
		Short: "Authenticate with Gitee",
		Use:   "gt auth login [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for login"},
			{Long: "--token", Type: "string", Description: "Gitee access token"},
		},
	},
	"auth logout": {
		Short: "Log out of Gitee",
		Use:   "gt auth logout",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for logout"},
		},
	},
	"auth setup-git": {
		Short: "Configure git to use Gitee CLI for credentials",
		Use:   "gt auth setup-git",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for setup-git"},
		},
	},
	"auth status": {
		Short: "View authentication status",
		Use:   "gt auth status",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for status"},
		},
	},
	"auth token": {
		Short: "Print current access token",
		Use:   "gt auth token",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for token"},
		},
	},
	"commit": {
		Short: "Run `git commit`",
		Use:   "gt commit [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for commit"},
		},
	},
	"completion": {
		Short: "Generate shell completion scripts",
		Use:   "gt completion [bash|zsh|fish|powershell]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for completion"},
		},
	},
	"config": {
		Short: "Manage gitee CLI config",
		Use:   "gt config <command>",
		Commands: []helpCommand{
			{Name: "get", Description: "Get configuration value"},
			{Name: "list", Description: "List all configuration settings"},
			{Name: "path", Description: "Print path to configuration file"},
			{Name: "set", Description: "Set configuration value"},
			{Name: "unset", Description: "Unset configuration value"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for config"},
		},
	},
	"config get": {
		Short: "Get configuration value",
		Use:   "gt config get <key>",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for get"},
		},
	},
	"config list": {
		Short: "List all configuration settings",
		Use:   "gt config list",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for list"},
		},
	},
	"config path": {
		Short: "Print path to configuration file",
		Use:   "gt config path",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for path"},
		},
	},
	"config set": {
		Short: "Set configuration value",
		Use:   "gt config set <key> <value>",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for set"},
		},
	},
	"config unset": {
		Short: "Unset configuration value",
		Use:   "gt config unset <key>",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for unset"},
		},
	},
	"git": {
		Short: "Run git operations with retry for transient network failures",
		Use:   "gt git <command>",
		Commands: []helpCommand{
			{Name: "commit", Description: "Run `git commit`"},
			{Name: "pull", Description: "Run `git pull`"},
			{Name: "push", Description: "Run `git push`"},
			{Name: "status", Description: "Run `git status`"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for git"},
		},
	},
	"git commit": {
		Short: "Run `git commit`",
		Use:   "gt git commit [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for commit"},
		},
	},
	"git pull": {
		Short: "Run `git pull`",
		Use:   "gt git pull [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for pull"},
		},
	},
	"git push": {
		Short: "Run `git push`",
		Use:   "gt git push [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for push"},
		},
	},
	"git status": {
		Short: "Run `git status`",
		Use:   "gt git status [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for status"},
		},
	},
	"help": {
		Short: "Help about any command",
		Use:   "gt help [command]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for help"},
		},
	},
	"issue": {
		Short: "Work with issues",
		Use:   "gt issue <command>",
		Commands: []helpCommand{
			{Name: "close", Description: "Close an issue"},
			{Name: "comment", Description: "Add a comment to an issue"},
			{Name: "create", Description: "Create an issue"},
			{Name: "list", Description: "List issues in a repository"},
			{Name: "reopen", Description: "Reopen an issue"},
			{Name: "status", Description: "Show open issues assigned to or created by you"},
			{Name: "view", Description: "View an issue"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for issue"},
		},
	},
	"issue close": {
		Short: "Close an issue",
		Use:   "gt issue close <number> [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for close"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"issue comment": {
		Short: "Add a comment to an issue",
		Use:   "gt issue comment <number> [flags]",
		Flags: []helpFlag{
			{Long: "--body", Type: "string", Description: "Comment body"},
			{Short: "-h", Long: "--help", Description: "help for comment"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"issue create": {
		Short: "Create an issue",
		Use:   "gt issue create [flags]",
		Flags: []helpFlag{
			{Long: "--body", Type: "string", Description: "Issue body"},
			{Short: "-h", Long: "--help", Description: "help for create"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
			{Long: "--title", Type: "string", Description: "Issue title"},
		},
	},
	"issue list": {
		Short: "List issues in a repository",
		Use:   "gt issue list [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for list"},
			{Long: "--page", Type: "int", Description: "Page number (default 1)"},
			{Long: "--per-page", Type: "int", Description: "Page size (default 30)"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
			{Long: "--state", Type: "string", Description: "open|closed|all (default \"open\")"},
		},
	},
	"issue reopen": {
		Short: "Reopen an issue",
		Use:   "gt issue reopen <number> [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for reopen"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"issue status": {
		Short: "Show open issues assigned to or created by you",
		Use:   "gt issue status",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for status"},
		},
	},
	"issue view": {
		Short: "View an issue",
		Use:   "gt issue view <number> [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for view"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"pr": {
		Short: "Work with pull requests",
		Use:   "gt pr <command>",
		Commands: []helpCommand{
			{Name: "close", Description: "Close a pull request"},
			{Name: "comment", Description: "Add a comment to a pull request"},
			{Name: "create", Description: "Create a pull request"},
			{Name: "list", Description: "List pull requests"},
			{Name: "merge", Description: "Merge a pull request"},
			{Name: "view", Description: "View a pull request"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for pr"},
		},
	},
	"pr close": {
		Short: "Close a pull request",
		Use:   "gt pr close [<number>] [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for close"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"pr comment": {
		Short: "Add a comment to a pull request",
		Use:   "gt pr comment [<number>] [flags]",
		Flags: []helpFlag{
			{Long: "--body", Type: "string", Description: "Comment body"},
			{Short: "-h", Long: "--help", Description: "help for comment"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"pr create": {
		Short: "Create a pull request",
		Use:   "gt pr create [flags]",
		Flags: []helpFlag{
			{Long: "--base", Type: "string", Description: "Base branch"},
			{Long: "--body", Type: "string", Description: "PR body"},
			{Long: "--head", Type: "string", Description: "Head branch"},
			{Short: "-h", Long: "--help", Description: "help for create"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
			{Long: "--title", Type: "string", Description: "PR title"},
		},
	},
	"pr list": {
		Short: "List pull requests",
		Use:   "gt pr list [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for list"},
			{Long: "--page", Type: "int", Description: "Page number (default 1)"},
			{Long: "--per-page", Type: "int", Description: "Page size (default 30)"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
			{Long: "--state", Type: "string", Description: "open|closed|all (default \"open\")"},
		},
	},
	"pr merge": {
		Short: "Merge a pull request",
		Use:   "gt pr merge [<number>] [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for merge"},
			{Long: "--message", Type: "string", Description: "Merge message"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"pr view": {
		Short: "View a pull request",
		Use:   "gt pr view [<number>] [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for view"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"pull": {
		Short: "Run `git pull`",
		Use:   "gt pull [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for pull"},
		},
	},
	"push": {
		Short: "Run `git push`",
		Use:   "gt push [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for push"},
		},
	},
	"release": {
		Short: "Manage releases",
		Use:   "gt release <command>",
		Commands: []helpCommand{
			{Name: "create", Description: "Create a release"},
			{Name: "delete", Description: "Delete a release"},
			{Name: "list", Description: "List releases in a repository"},
			{Name: "view", Description: "View a release"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for release"},
		},
	},
	"release create": {
		Short: "Create a release",
		Use:   "gt release create [flags]",
		Flags: []helpFlag{
			{Long: "--body", Type: "string", Description: "Release notes"},
			{Long: "--draft", Description: "Create as draft"},
			{Short: "-h", Long: "--help", Description: "help for create"},
			{Long: "--name", Type: "string", Description: "Release name"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
			{Long: "--tag", Type: "string", Description: "Tag name"},
			{Long: "--target", Type: "string", Description: "Target commitish"},
		},
	},
	"release delete": {
		Short: "Delete a release",
		Use:   "gt release delete <id> [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for delete"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"release list": {
		Short: "List releases in a repository",
		Use:   "gt release list [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for list"},
			{Long: "--page", Type: "int", Description: "Page number (default 1)"},
			{Long: "--per-page", Type: "int", Description: "Page size (default 30)"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"release view": {
		Short: "View a release",
		Use:   "gt release view <tag> [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for view"},
			{Long: "--repo", Type: "string", Description: "Repository owner/name"},
		},
	},
	"repo": {
		Short: "Work with repositories",
		Use:   "gt repo <command>",
		Commands: []helpCommand{
			{Name: "clone", Description: "Clone a repository"},
			{Name: "create", Description: "Create a new repository"},
			{Name: "list", Description: "List repositories"},
			{Name: "view", Description: "View a repository"},
		},
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for repo"},
		},
	},
	"repo clone": {
		Short: "Clone a repository",
		Use:   "gt repo clone <repo> [dest] [flags]",
		Flags: []helpFlag{
			{Long: "--depth", Type: "int", Description: "Create a shallow clone"},
			{Long: "--dest", Type: "string", Description: "Destination directory"},
			{Short: "-h", Long: "--help", Description: "help for clone"},
			{Long: "--recursive", Description: "Clone submodules"},
			{Long: "--ssh", Description: "Use SSH protocol for cloning"},
		},
	},
	"repo create": {
		Short: "Create a new repository",
		Use:   "gt repo create [flags]",
		Flags: []helpFlag{
			{Long: "--description", Type: "string", Description: "Repository description"},
			{Short: "-h", Long: "--help", Description: "help for create"},
			{Long: "--name", Type: "string", Description: "Repository name"},
			{Long: "--org", Type: "string", Description: "Organization name"},
			{Long: "--private", Description: "Create private repository"},
		},
	},
	"repo list": {
		Short: "List repositories",
		Use:   "gt repo list [flags]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for list"},
			{Long: "--org", Type: "string", Description: "Organization name"},
			{Long: "--page", Type: "int", Description: "Page number (default 1)"},
			{Long: "--per-page", Type: "int", Description: "Page size (default 30)"},
			{Long: "--visibility", Type: "string", Description: "all|public|private (default \"all\")"},
		},
	},
	"repo view": {
		Short: "View a repository",
		Use:   "gt repo view [<repo>]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for view"},
		},
	},
	"status": {
		Short: "Run `git status`",
		Use:   "gt status [git args...]",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for status"},
		},
	},
	"version": {
		Short: "Print version",
		Use:   "gt version",
		Flags: []helpFlag{
			{Short: "-h", Long: "--help", Description: "help for version"},
		},
	},
}
