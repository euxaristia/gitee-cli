package cmd

import (
	"os"
	"testing"
)

func TestPRList(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("pr list error = %v", err)
	}
}

func TestPRView(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"view", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr view error = %v", err)
	}
}

func TestPRView_JSON(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.Output = "json"

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"view", "1", "--repo", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("pr view json error = %v", err)
	}
}

func TestPRView_InvalidNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"view", "abc", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr view with invalid number expected error")
	}
}

func TestPRCreate(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--title", "My PR", "--head", "feature", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr create error = %v", err)
	}
}

func TestPRCreate_JSON(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.Output = "json"

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--title", "My PR", "--head", "feature", "--base", "main"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("pr create json error = %v", err)
	}
}

func TestPRCreate_MissingFlags(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr create without required flags expected error")
	}
}

func TestPRMerge(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"merge", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr merge error = %v", err)
	}
}

func TestPRMerge_ImplicitNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"merge", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr merge without number error = %v", err)
	}
}

func TestPRMerge_InvalidNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"merge", "abc", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr merge with invalid number expected error")
	}
}

func TestPRClose(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"close", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr close error = %v", err)
	}
}

func TestPRClose_ImplicitNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"close", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr close without number error = %v", err)
	}
}

func TestPRClose_InvalidNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"close", "abc", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr close with invalid number expected error")
	}
}

func TestPRComment(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"comment", "1", "--repo", "owner/repo", "--body", "Nice!"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr comment error = %v", err)
	}
}

func TestPRComment_ImplicitNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"comment", "--repo", "owner/repo", "--body", "Nice!"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("pr comment without number error = %v", err)
	}
}

func TestPRComment_NoBody(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"comment", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr comment without body expected error")
	}
}

func TestPRComment_InvalidNumber(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"comment", "abc", "--repo", "owner/repo", "--body", "test"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr comment with invalid number expected error")
	}
}

func TestPRList_NoToken(t *testing.T) {
	app := testAppNoToken()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("pr list without token expected error")
	}
}

func TestPRList_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRView_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"view", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRCreate_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--title", "PR", "--head", "f", "--base", "m"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRMerge_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"merge", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRClose_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"close", "1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRComment_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"comment", "1", "--repo", "owner/repo", "--body", "test"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRList_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRView_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"view", "1", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRCreate_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "badrepo", "--title", "PR", "--head", "f", "--base", "m"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRMerge_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"merge", "1", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRClose_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"close", "1", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestPRComment_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newPRCmd(app)
	cmd.SetArgs([]string{"comment", "1", "--repo", "badrepo", "--body", "test"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}
