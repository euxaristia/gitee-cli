package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const (
	gitRetryAttempts = 3
	gitRetryBaseWait = 1 * time.Second
)

type gitRunner func(context.Context, []string, io.Reader, io.Writer, io.Writer) error

func runGitCommand(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	c := exec.CommandContext(ctx, "git", args...)
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}

func appGitRunner(app *App) gitRunner {
	if app != nil && app.GitRunner != nil {
		return app.GitRunner
	}
	return runGitCommand
}

func appContext(app *App) context.Context {
	if app != nil && app.Ctx != nil {
		return app.Ctx
	}
	return context.Background()
}

func newGitCmd(app *App) *Command {
	return &Command{
		Use:   "git",
		Short: "Run git operations with retry for transient network failures",
		run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("git requires a subcommand")
			}
			if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
				return runHelp(app, c, append([]string{"git"}, args[1:]...))
			}
			switch args[0] {
			case "commit", "push", "pull", "status":
				if len(args) > 1 && (args[1] == "help" || args[1] == "--help" || args[1] == "-h") {
					return runHelp(app, c, []string{"git", args[0]})
				}
				return runGitWithRetry(app, c, args[0], args[1:])
			default:
				return fmt.Errorf("unknown git command %q", args[0])
			}
		},
	}
}

func newGitShortcutCmd(app *App, name string) *Command {
	return newGitOperationCmd(app, name)
}

func newGitOperationCmd(app *App, op string) *Command {
	return &Command{
		Use:   fmt.Sprintf("%s [git args...]", op),
		Short: fmt.Sprintf("Run `git %s`", op),
		run: func(c *Command, args []string) error {
			if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
				return runHelp(app, c, []string{op})
			}
			return runGitWithRetry(app, c, op, args)
		},
	}
}

func runGitWithRetry(app *App, cmd *Command, op string, args []string) error {
	attempts := 1
	if op == "push" || op == "pull" || op == "clone" {
		attempts = gitRetryAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		var stdoutBuf bytes.Buffer
		var stderrBuf bytes.Buffer

		multiStdout := io.MultiWriter(cmd.OutOrStdout(), &stdoutBuf)
		multiStderr := io.MultiWriter(cmd.ErrOrStderr(), &stderrBuf)

		gitArgs := []string{
			"-c", "protocol.version=2",
			"-c", "http.version=HTTP/1.1",
			"-c", "http.postBuffer=524288000",
		}
		if app != nil && app.Cfg != nil && len(app.Cfg.GitFlags) > 0 {
			gitArgs = append(gitArgs, app.Cfg.GitFlags...)
		}
		gitArgs = append(gitArgs, op)
		gitArgs = append(gitArgs, args...)

		err := appGitRunner(app)(appContext(app), gitArgs, cmd.InOrStdin(), multiStdout, multiStderr)
		if err == nil {
			if attempt > 1 {
				fmt.Fprintf(cmd.ErrOrStderr(), "gt: git %s succeeded on retry %d/%d\n", op, attempt, attempts)
			}
			return nil
		}
		lastErr = err

		errText := strings.ToLower(stderrBuf.String() + "\n" + stdoutBuf.String() + "\n" + err.Error())
		if attempt == attempts || !isTransientGitErr(errText) {
			break
		}

		wait := time.Duration(attempt) * gitRetryBaseWait
		fmt.Fprintf(cmd.ErrOrStderr(), "gt: transient failure running git %s (attempt %d/%d), retrying in %s\n", op, attempt, attempts, wait)
		time.Sleep(wait)
	}

	return fmt.Errorf("git %s failed: %w", op, lastErr)
}

func isTransientGitErr(msg string) bool {
	return strings.Contains(msg, "tls handshake timeout") ||
		strings.Contains(msg, "gnutls_handshake() failed") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "unexpected eof") ||
		strings.Contains(msg, "remote end hung up unexpectedly") ||
		strings.Contains(msg, "ssh_exchange_identification: read: connection reset by peer") ||
		strings.Contains(msg, "kex_exchange_identification: read: connection reset by peer") ||
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
