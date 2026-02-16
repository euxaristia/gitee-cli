package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	gitRetryAttempts = 3
	gitRetryBaseWait = 1 * time.Second
)

func newGitCmd(_ *App) *cobra.Command {
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Run git operations with retry for transient network failures",
	}

	gitCmd.AddCommand(
		newGitOperationCmd("commit"),
		newGitOperationCmd("push"),
		newGitOperationCmd("pull"),
	)
	return gitCmd
}

func newGitShortcutCmd(name string) *cobra.Command {
	cmd := newGitOperationCmd(name)
	cmd.Use = fmt.Sprintf("%s [git args...]", name)
	return cmd
}

func newGitOperationCmd(op string) *cobra.Command {
	return &cobra.Command{
		Use:                fmt.Sprintf("%s [git args...]", op),
		Short:              fmt.Sprintf("Run `git %s`", op),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGitWithRetry(cmd, op, args)
		},
	}
}

func runGitWithRetry(cmd *cobra.Command, op string, args []string) error {
	attempts := 1
	if op == "push" || op == "pull" {
		attempts = gitRetryAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		var stdoutBuf bytes.Buffer
		var stderrBuf bytes.Buffer

		gitArgs := append([]string{op}, args...)
		c := exec.Command("git", gitArgs...)
		c.Stdin = os.Stdin
		c.Stdout = &stdoutBuf
		c.Stderr = &stderrBuf

		err := c.Run()
		if err == nil {
			_, _ = io.Copy(cmd.OutOrStdout(), &stdoutBuf)
			_, _ = io.Copy(cmd.ErrOrStderr(), &stderrBuf)
			if attempt > 1 {
				fmt.Fprintf(cmd.ErrOrStderr(), "gitee: git %s succeeded on retry %d/%d\n", op, attempt, attempts)
			}
			return nil
		}

		errText := strings.ToLower(stderrBuf.String() + "\n" + stdoutBuf.String() + "\n" + err.Error())
		if attempt == attempts || !isTransientGitErr(errText) {
			_, _ = io.Copy(cmd.OutOrStdout(), &stdoutBuf)
			_, _ = io.Copy(cmd.ErrOrStderr(), &stderrBuf)
			return fmt.Errorf("git %s failed: %w", op, err)
		}

		wait := time.Duration(attempt) * gitRetryBaseWait
		fmt.Fprintf(cmd.ErrOrStderr(), "gitee: transient failure running git %s (attempt %d/%d), retrying in %s\n", op, attempt, attempts, wait)
		time.Sleep(wait)
	}

	return nil
}

func isTransientGitErr(msg string) bool {
	return strings.Contains(msg, "tls handshake timeout") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "unexpected eof") ||
		strings.Contains(msg, "remote end hung up unexpectedly") ||
		strings.Contains(msg, "operation timed out") ||
		strings.Contains(msg, "connection timed out") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "could not resolve host") ||
		strings.Contains(msg, "http 502") ||
		strings.Contains(msg, "http 503") ||
		strings.Contains(msg, "http 504") ||
		strings.Contains(msg, "the requested url returned error: 5") ||
		strings.Contains(msg, "failure when receiving data from the peer") ||
		strings.Contains(msg, "rpc failed")
}
