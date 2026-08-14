package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpCmd_Root(t *testing.T) {
	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("gt help error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "A full-featured CLI for Gitee") {
		t.Errorf("expected root help description, got:\n%s", out)
	}
	if !strings.Contains(out, "help        Help about any command") {
		t.Errorf("expected help command in available commands, got:\n%s", out)
	}
	if !strings.Contains(out, "repo        Work with repositories") {
		t.Errorf("expected repo command in available commands, got:\n%s", out)
	}
}

func TestHelpCmd_RootHelpFlags(t *testing.T) {
	tests := [][]string{
		{"--help"},
		{"-h"},
	}
	for _, args := range tests {
		cmd := NewRootCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("gt %v error = %v", args, err)
		}
		if !strings.Contains(buf.String(), "A full-featured CLI for Gitee") {
			t.Errorf("args %v did not output root help, got:\n%s", args, buf.String())
		}
	}
}

func TestHelpCmd_Subcommands(t *testing.T) {
	tests := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"help", "repo"},
			expected: []string{"Work with repositories", "gt repo <command>", "Available Commands:", "clone", "create", "list", "view"},
		},
		{
			args:     []string{"repo", "--help"},
			expected: []string{"Work with repositories", "gt repo <command>", "Available Commands:"},
		},
		{
			args:     []string{"repo", "-h"},
			expected: []string{"Work with repositories", "gt repo <command>"},
		},
		{
			args:     []string{"help", "repo", "list"},
			expected: []string{"List repositories", "gt repo list [flags]", "--org", "--page", "--per-page", "--visibility"},
		},
		{
			args:     []string{"repo", "list", "--help"},
			expected: []string{"List repositories", "gt repo list [flags]", "--org", "--page"},
		},
		{
			args:     []string{"help", "auth", "login"},
			expected: []string{"Authenticate with Gitee", "gt auth login [flags]", "--token"},
		},
		{
			args:     []string{"auth", "login", "--help"},
			expected: []string{"Authenticate with Gitee", "gt auth login [flags]", "--token"},
		},
		{
			args:     []string{"help", "api"},
			expected: []string{"Make a raw API request to Gitee v5", "gt api <endpoint> [flags]", "--field", "--header", "--method"},
		},
		{
			args:     []string{"api", "--help"},
			expected: []string{"Make a raw API request to Gitee v5", "gt api <endpoint> [flags]"},
		},
		{
			args:     []string{"help", "issue", "create"},
			expected: []string{"Create an issue", "gt issue create [flags]", "--title", "--body", "--repo"},
		},
		{
			args:     []string{"issue", "create", "--help"},
			expected: []string{"Create an issue", "gt issue create [flags]", "--title"},
		},
		{
			args:     []string{"help", "pr", "create"},
			expected: []string{"Create a pull request", "gt pr create [flags]", "--title", "--head", "--base"},
		},
		{
			args:     []string{"pr", "create", "--help"},
			expected: []string{"Create a pull request", "gt pr create [flags]"},
		},
		{
			args:     []string{"help", "release", "create"},
			expected: []string{"Create a release", "gt release create [flags]", "--tag", "--name"},
		},
		{
			args:     []string{"release", "create", "--help"},
			expected: []string{"Create a release", "gt release create [flags]"},
		},
		{
			args:     []string{"help", "config", "set"},
			expected: []string{"Set configuration value", "gt config set <key> <value>"},
		},
		{
			args:     []string{"config", "set", "--help"},
			expected: []string{"Set configuration value", "gt config set <key> <value>"},
		},
		{
			args:     []string{"help", "completion"},
			expected: []string{"Generate shell completion scripts", "gt completion [bash|zsh|fish|powershell]"},
		},
		{
			args:     []string{"completion", "--help"},
			expected: []string{"Generate shell completion scripts"},
		},
		{
			args:     []string{"help", "version"},
			expected: []string{"Print version", "gt version"},
		},
		{
			args:     []string{"version", "--help"},
			expected: []string{"Print version"},
		},
		{
			args:     []string{"help", "help"},
			expected: []string{"Help about any command", "gt help [command]"},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, "_"), func(t *testing.T) {
			cmd := NewRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute(%v) error = %v", tt.args, err)
			}
			out := buf.String()
			for _, exp := range tt.expected {
				if !strings.Contains(out, exp) {
					t.Errorf("args %v output missing %q, got:\n%s", tt.args, exp, out)
				}
			}
		})
	}
}

func TestHelpCmd_UnknownCommand(t *testing.T) {
	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"help", "unknown-command"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown help command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' in error, got: %v", err)
	}
}

func TestSubcommand_DirectHelpFlags(t *testing.T) {
	app := testAppNoToken()

	tests := []struct {
		name     string
		cmd      *Command
		args     []string
		expected string
	}{
		{"repo --help", newRepoCmd(app), []string{"--help"}, "Work with repositories"},
		{"repo list --help", newRepoCmd(app), []string{"list", "--help"}, "List repositories"},
		{"auth --help", newAuthCmd(app), []string{"--help"}, "Authenticate with Gitee"},
		{"auth login --help", newAuthCmd(app), []string{"login", "--help"}, "Authenticate with Gitee"},
		{"config --help", newConfigCmd(app), []string{"--help"}, "Manage gitee CLI config"},
		{"config get --help", newConfigCmd(app), []string{"get", "--help"}, "Get configuration value"},
		{"issue --help", newIssueCmd(app), []string{"--help"}, "Work with issues"},
		{"issue list --help", newIssueCmd(app), []string{"list", "--help"}, "List issues in a repository"},
		{"pr --help", newPRCmd(app), []string{"--help"}, "Work with pull requests"},
		{"pr list --help", newPRCmd(app), []string{"list", "--help"}, "List pull requests"},
		{"release --help", newReleaseCmd(app), []string{"--help"}, "Manage releases"},
		{"release list --help", newReleaseCmd(app), []string{"list", "--help"}, "List releases in a repository"},
		{"git --help", newGitCmd(app), []string{"--help"}, "Run git operations with retry"},
		{"api --help", newAPICmd(app), []string{"--help"}, "Make a raw API request"},
		{"version --help", newVersionCmd(), []string{"--help"}, "Print version"},
		{"completion --help", newCompletionCmd(), []string{"--help"}, "Generate shell completion scripts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.cmd.SetOut(&buf)
			tt.cmd.SetErr(&buf)
			tt.cmd.SetArgs(tt.args)
			if err := tt.cmd.Execute(); err != nil {
				t.Fatalf("%s error = %v", tt.name, err)
			}
			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("%s output missing %q, got:\n%s", tt.name, tt.expected, buf.String())
			}
		})
	}
}
