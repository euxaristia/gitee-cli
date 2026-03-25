package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCompletionCmd(t *testing.T) {
	root := &cobra.Command{Use: "gt"}

	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := newCompletionCmd(root)
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs([]string{shell})
			if err := cmd.Execute(); err != nil {
				t.Errorf("completion %s error = %v", shell, err)
			}
		})
	}
}

func TestNewCompletionCmd_Unsupported(t *testing.T) {
	root := &cobra.Command{Use: "gt"}
	cmd := newCompletionCmd(root)
	cmd.SetArgs([]string{"elvish"})
	if err := cmd.Execute(); err == nil {
		t.Error("completion elvish expected error")
	}
}
