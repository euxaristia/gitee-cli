package cmd

import (
	"os"
	"testing"
)

func TestReleaseList(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("release list error = %v", err)
	}
}

func TestReleaseView(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"view", "v1.0.0", "--repo", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("release view error = %v", err)
	}
}

func TestReleaseCreate(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--tag", "v1.0.0", "--name", "Release 1.0"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("release create error = %v", err)
	}
}

func TestReleaseCreate_NoTag(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("release create without tag expected error")
	}
}

func TestReleaseDelete(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"delete", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("release delete error = %v", err)
	}
}

func TestReleaseDelete_InvalidID(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"delete", "abc", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("release delete with invalid id expected error")
	}
}

func TestReleaseList_NoToken(t *testing.T) {
	app := testAppNoToken()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("release list without token expected error")
	}
}

func TestReleaseList_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseView_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"view", "v1.0", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseCreate_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--tag", "v1.0"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseDelete_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"delete", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseList_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseView_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"view", "v1.0", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseCreate_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "badrepo", "--tag", "v1.0"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReleaseDelete_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newReleaseCmd(app)
	cmd.SetArgs([]string{"delete", "1", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}
