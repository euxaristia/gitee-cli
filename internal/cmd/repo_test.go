package cmd

import (
	"os"
	"testing"
)

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

	cmd := newRepoCmd(app)
	// Clone a non-existent repo - will fail but exercises the SSH URL branch
	cmd.SetArgs([]string{"clone", "owner/nonexistent-test-repo-12345", "--ssh", "--dest", t.TempDir()})
	// Expected to fail since repo doesn't exist on gitee.com
	_ = cmd.Execute()
}

func TestRepoClone_WithOptions(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newRepoCmd(app)
	cmd.SetArgs([]string{"clone", "owner/nonexistent-test-repo-12345", "--depth", "1", "--recursive", "--dest", t.TempDir()})
	_ = cmd.Execute()
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
