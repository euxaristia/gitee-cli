package cmd

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	keyring "github.com/zalando/go-keyring"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/auth"
	"github.com/euxaristia/gitee-cli/internal/config"
)

func init() {
	keyring.MockInit()
}

func TestAuthLogin_WithToken(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	tmp := t.TempDir()
	useConfigDir(t, tmp)

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"login", "--token", "test-token-123"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth login error = %v", err)
	}
}

func TestAuthLogin_EmptyToken(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	tmp := t.TempDir()
	useConfigDir(t, tmp)

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"login", "--token", ""})
	// This will try ReadTokenFromTTY which fails on non-terminal
	if err := cmd.Execute(); err == nil {
		t.Error("auth login with empty token expected error")
	}
}

func TestAuthLogout(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	tmp := t.TempDir()
	useConfigDir(t, tmp)

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"logout"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth logout error = %v", err)
	}
}

func TestAuthStatus_WithToken(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	var buf bytes.Buffer
	cmd := newAuthCmd(app)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth status error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "https://gitee.com") {
		t.Errorf("expected host in output, got %q", out)
	}
	if !strings.Contains(out, "✓ Logged in") {
		t.Errorf("expected checkmark logged in, got %q", out)
	}
	if !strings.Contains(out, "testuser") {
		t.Errorf("expected username in output, got %q", out)
	}
	if !strings.Contains(out, "Active account: true") {
		t.Errorf("expected active account, got %q", out)
	}
	if !strings.Contains(out, "Git operations protocol: https") {
		t.Errorf("expected git protocol, got %q", out)
	}
	if !strings.Contains(out, "Token:") {
		t.Errorf("expected masked token, got %q", out)
	}
	// Token should be masked — full token must not appear
	if strings.Contains(out, "test-token") {
		t.Errorf("full token should not appear in output, got %q", out)
	}
}

func TestAuthStatus_NoToken(t *testing.T) {
	app := testAppNoToken()
	var buf bytes.Buffer
	cmd := newAuthCmd(app)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth status no token error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "https://gitee.com") {
		t.Errorf("expected host in output, got %q", out)
	}
	if !strings.Contains(out, "Not authenticated") {
		t.Errorf("expected not authenticated message, got %q", out)
	}
}

func TestAuthStatus_WithEnvToken(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	t.Setenv("GITEE_TOKEN", "env-token-value")

	var buf bytes.Buffer
	cmd := newAuthCmd(app)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth status with env token error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "GITEE_TOKEN") {
		t.Errorf("expected GITEE_TOKEN source, got %q", out)
	}
}

func TestAuthStatus_WithKeychainToken(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	t.Setenv("GITEE_TOKEN", "")
	// Save a token to keyring so tokenSource returns "keyring"
	if err := auth.SaveToken("keyring-test-token"); err != nil {
		t.Fatalf("failed to save keyring token: %v", err)
	}
	defer auth.DeleteToken()

	var buf bytes.Buffer
	cmd := newAuthCmd(app)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth status with keychain error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "keyring") {
		t.Errorf("expected keyring source, got %q", out)
	}
}

func TestAuthToken_FromEnv(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	t.Setenv("GITEE_TOKEN", "env-token")
	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"token"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth token error = %v", err)
	}
}

func TestAuthToken_FromKeychain(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	t.Setenv("GITEE_TOKEN", "")
	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"token"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth token keychain error = %v", err)
	}
}

func TestAuthToken_None(t *testing.T) {
	app := testAppNoToken()
	t.Setenv("GITEE_TOKEN", "")
	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"token"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth token none error = %v", err)
	}
}

func TestAuthGitCredential_NoArgs(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential no args error = %v", err)
	}
}

func TestAuthGitCredential_NotGet(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential", "store"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential store error = %v", err)
	}
}

func TestAuthGitCredential_Get(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.User = "testuser"

	// Feed stdin with host=gitee.com
	r, w, _ := os.Pipe()
	w.WriteString("host=gitee.com\n\n")
	w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential", "get"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential get error = %v", err)
	}
}

func TestAuthGitCredential_Get_NoUser(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.User = ""

	r, w, _ := os.Pipe()
	w.WriteString("host=gitee.com\n\n")
	w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential", "get"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential get no user error = %v", err)
	}
}

func TestAuthGitCredential_Get_WrongHost(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	w.WriteString("host=github.com\n\n")
	w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential", "get"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential wrong host error = %v", err)
	}
}

func TestAuthGitCredential_Get_NoToken(t *testing.T) {
	app := testAppNoToken()

	r, w, _ := os.Pipe()
	w.WriteString("host=gitee.com\n\n")
	w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential", "get"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential no token error = %v", err)
	}
}

func TestClientFrom(t *testing.T) {
	app := &App{
		Cfg: config.Default(),
		Ctx: context.Background(),
	}
	client := clientFrom(app, "test-token")
	if client == nil {
		t.Error("clientFrom returned nil")
	}
}

func TestEnsureToken_WithEnv(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "env-token")
	app := &App{
		Cfg:    config.Default(),
		Client: api.New(config.Default().APIBase, ""),
	}
	if err := ensureToken(app); err != nil {
		t.Errorf("ensureToken with env error = %v", err)
	}
	if app.ActiveToken != "env-token" {
		t.Errorf("ActiveToken = %q, want env-token", app.ActiveToken)
	}
}

func TestEnsureToken_NoToken(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "")
	app := testAppNoToken()
	err := ensureToken(app)
	if err == nil {
		t.Error("ensureToken without token expected error")
	}
}

func TestAuthGitCredential_Get_NoUser_APIError(t *testing.T) {
	// Test git-credential get when User is empty and CurrentUser API fails
	cfg := config.Default()
	cfg.APIBase = "http://127.0.0.1:1" // unreachable
	app := &App{
		Cfg:         cfg,
		Client:      api.New("http://127.0.0.1:1", "test-token"),
		ActiveToken: "test-token",
		Ctx:         context.Background(),
	}

	r, w, _ := os.Pipe()
	w.WriteString("host=gitee.com\n\n")
	w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"git-credential", "get"})
	// Should fall back to "oauth2" username
	if err := cmd.Execute(); err != nil {
		t.Errorf("auth git-credential get API error = %v", err)
	}
}

func TestAuthStatus_APIError(t *testing.T) {
	app := testErrorApp()
	var buf bytes.Buffer
	cmd := newAuthCmd(app)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err == nil {
		t.Error("auth status with API error expected error")
	}
	out := buf.String()
	if !strings.Contains(out, "Failed to verify token") {
		t.Errorf("expected failure message, got %q", out)
	}
}

func TestAuthSetupGit(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAuthCmd(app)
	cmd.SetArgs([]string{"setup-git"})
	// This will run git config --global, which may or may not succeed
	// depending on git being available. We just exercise the code path.
	_ = cmd.Execute()
}

func TestEnsureToken_WithActiveToken(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "")
	app := &App{
		Cfg:         config.Default(),
		ActiveToken: "existing-token",
	}
	if err := ensureToken(app); err != nil {
		t.Errorf("ensureToken with active token error = %v", err)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"ab", "**"},
		{"abcd", "****"},
		{"abcde", "abcd*"},
		{"abcdefghij", "abcd******"},
		{"gho_ABCDEFGHIJKLMNOPQRSTUVWXYZab", "gho_****************************"},
	}
	for _, tt := range tests {
		got := maskToken(tt.input)
		if got != tt.want {
			t.Errorf("maskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTokenSource_Env(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "env-token")
	app := &App{Cfg: config.Default()}
	if got := tokenSource(app); got != "GITEE_TOKEN" {
		t.Errorf("tokenSource = %q, want GITEE_TOKEN", got)
	}
}

func TestTokenSource_Keyring(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "")
	if err := auth.SaveToken("keyring-token"); err != nil {
		t.Fatalf("failed to save keyring token: %v", err)
	}
	defer auth.DeleteToken()

	app := &App{Cfg: config.Default()}
	if got := tokenSource(app); got != "keyring" {
		t.Errorf("tokenSource = %q, want keyring", got)
	}
}

func TestTokenSource_ConfigFile(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "")
	// Make sure keyring is empty
	auth.DeleteToken()

	cfg := config.Default()
	cfg.Token = "legacy-token"
	app := &App{Cfg: cfg}
	if got := tokenSource(app); got != "config file" {
		t.Errorf("tokenSource = %q, want config file", got)
	}
}

func TestTokenSource_Unknown(t *testing.T) {
	t.Setenv("GITEE_TOKEN", "")
	auth.DeleteToken()

	app := &App{Cfg: config.Default()}
	if got := tokenSource(app); got != "unknown" {
		t.Errorf("tokenSource = %q, want unknown", got)
	}
}
