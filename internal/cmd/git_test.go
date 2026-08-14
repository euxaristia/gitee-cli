package cmd

import (
	"bytes"
	"testing"

	"github.com/euxaristia/gitee-cli/internal/config"
)

func TestIsTransientGitErr(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"tls handshake timeout", true},
		{"gnutls_handshake() failed", true},
		{"connection reset by peer", true},
		{"unexpected eof", true},
		{"remote end hung up unexpectedly", true},
		{"ssh_exchange_identification: read: connection reset by peer", true},
		{"kex_exchange_identification: read: connection reset by peer", true},
		{"operation timed out", true},
		{"connection timed out", true},
		{"network is unreachable", true},
		{"could not resolve host", true},
		{"http 502", true},
		{"http 503", true},
		{"http 504", true},
		{"the requested url returned error: 5", true},
		{"failure when receiving data from the peer", true},
		{"rpc failed", true},
		{"normal error", false},
		{"exit status 1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			if got := isTransientGitErr(tt.msg); got != tt.want {
				t.Errorf("isTransientGitErr(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestRunGitWithRetry_Success(t *testing.T) {
	app := &App{Cfg: config.Default()}
	cmd := &Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := runGitWithRetry(app, cmd, "version", nil)
	if err != nil {
		t.Errorf("runGitWithRetry(version) error = %v", err)
	}
}

func TestRunGitWithRetry_WithGitFlags(t *testing.T) {
	cfg := config.Default()
	cfg.GitFlags = []string{"--no-pager"}
	app := &App{Cfg: cfg}
	cmd := &Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := runGitWithRetry(app, cmd, "version", nil)
	if err != nil {
		t.Errorf("runGitWithRetry(version with flags) error = %v", err)
	}
}

func TestRunGitWithRetry_Failure(t *testing.T) {
	app := &App{Cfg: config.Default()}
	cmd := &Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := runGitWithRetry(app, cmd, "nonexistentcommand", nil)
	if err == nil {
		t.Error("runGitWithRetry(nonexistentcommand) expected error")
	}
}

func TestRunGitWithRetry_NilApp(t *testing.T) {
	cmd := &Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := runGitWithRetry(nil, cmd, "version", nil)
	if err != nil {
		t.Errorf("runGitWithRetry(nil app) error = %v", err)
	}
}

func TestNewGitCmd(t *testing.T) {
	app := &App{Cfg: config.Default()}
	cmd := newGitCmd(app)
	if cmd.Use != "git" {
		t.Errorf("newGitCmd().Use = %q, want git", cmd.Use)
	}
}

func TestNewGitShortcutCmd(t *testing.T) {
	app := &App{Cfg: config.Default()}
	cmd := newGitShortcutCmd(app, "push")
	if cmd.Use != "push [git args...]" {
		t.Errorf("newGitShortcutCmd(push).Use = %q", cmd.Use)
	}
}

func TestNewGitOperationCmd(t *testing.T) {
	app := &App{Cfg: config.Default()}
	cmd := newGitOperationCmd(app, "status")
	if cmd.Use != "status [git args...]" {
		t.Errorf("newGitOperationCmd(status).Use = %q", cmd.Use)
	}
}
