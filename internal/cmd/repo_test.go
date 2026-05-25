package cmd

import (
	"context"
	"io"
	"os"
	"reflect"
	"testing"
)

func captureGitArgs(app *App) *[]string {
	var gotArgs []string
	app.GitRunner = func(_ context.Context, args []string, _ io.Reader, _, _ io.Writer) error {
		gotArgs = append([]string(nil), args...)
		return nil
	}
	return &gotArgs
}

func TestRepoList(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("repo list error = %v", err)
	}
}

func TestRepoView(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"view", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("repo view error = %v", err)
	}
}

func TestRepoView_InvalidRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"view", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("repo view with bad repo expected error")
	}
}

func TestRepoCreate(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"create", "--name", "newrepo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("repo create error = %v", err)
	}
}

func TestRepoCreate_JSON(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.Output = "json"

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"create", "--name", "newrepo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("repo create json error = %v", err)
	}
}

func TestRepoCreate_NoName(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"create"})
	if err := cmd.Execute(); err == nil {
		t.Error("repo create without name expected error")
	}
}

func TestRepoList_NoToken(t *testing.T) {
	app := testAppNoToken()
	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err == nil {
		t.Error("repo list without token expected error")
	}
}

func TestRepoClone_InvalidRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"clone", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("repo clone with bad repo expected error")
	}
}

func TestRepoClone_SSH(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	gotArgs := captureGitArgs(app)

	cmd := newRepoCmd(app)
	dest := t.TempDir()
	cmd.SetArgs([]string{"clone", "owner/nonexistent-test-repo-12345", "--ssh", "--dest", dest})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("repo clone ssh error = %v", err)
	}

	wantArgs := []string{
		"-c", "protocol.version=2",
		"-c", "http.version=HTTP/1.1",
		"-c", "http.postBuffer=524288000",
		"clone", "git@gitee.com:owner/nonexistent-test-repo-12345.git", dest,
	}
	if !reflect.DeepEqual(*gotArgs, wantArgs) {
		t.Errorf("git args = %v, want %v", *gotArgs, wantArgs)
	}
}

func TestRepoClone_WithOptions(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	gotArgs := captureGitArgs(app)

	cmd := newRepoCmd(app)
	dest := t.TempDir()
	cmd.SetArgs([]string{"clone", "owner/nonexistent-test-repo-12345", "--depth", "1", "--recursive", "--dest", dest})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("repo clone with options error = %v", err)
	}

	wantArgs := []string{
		"-c", "protocol.version=2",
		"-c", "http.version=HTTP/1.1",
		"-c", "http.postBuffer=524288000",
		"clone", "https://gitee.com/owner/nonexistent-test-repo-12345.git", "--depth", "1", "--recursive", dest,
	}
	if !reflect.DeepEqual(*gotArgs, wantArgs) {
		t.Errorf("git args = %v, want %v", *gotArgs, wantArgs)
	}
}

func TestRepoList_WithOrg(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"list", "--org", "myorg"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("repo list with org error = %v", err)
	}
}

func TestRepoList_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestRepoView_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"view", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestRepoCreate_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"create", "--name", "repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestRepoCreate_Private(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"create", "--name", "newrepo", "--private"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("error = %v", err)
	}
}

func TestRepoCreate_WithOrg(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"create", "--name", "newrepo", "--org", "myorg"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("repo create with org error = %v", err)
	}
}
