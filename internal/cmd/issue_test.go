package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/config"
)

func TestIssueList(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("issue list error = %v", err)
	}
}

func TestIssueView(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"view", "I1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("issue view error = %v", err)
	}
}

func TestIssueView_JSON(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.Output = "json"

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"view", "I1", "--repo", "owner/repo"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("issue view json error = %v", err)
	}
}

func TestIssueCreate(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--title", "New Bug"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("issue create error = %v", err)
	}
}

func TestIssueCreate_JSON(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	app.Cfg.Output = "json"

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--title", "New Bug"})
	err := cmd.Execute()

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("issue create json error = %v", err)
	}
}

func TestIssueCreate_NoTitle(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue create without title expected error")
	}
}

func TestIssueComment(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"comment", "I1", "--repo", "owner/repo", "--body", "A comment"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("issue comment error = %v", err)
	}
}

func TestIssueComment_NoBody(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"comment", "I1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue comment without body expected error")
	}
}

func TestIssueClose(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"close", "I1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("issue close error = %v", err)
	}
}

func TestIssueReopen(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"reopen", "I1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("issue reopen error = %v", err)
	}
}

func TestIssueStatus(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("issue status error = %v", err)
	}
}

func TestIssueList_NoToken(t *testing.T) {
	app := testAppNoToken()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue list without token expected error")
	}
}

func TestChangeIssueState_NoToken(t *testing.T) {
	app := testAppNoToken()
	err := changeIssueState(app, "owner/repo", "I1", "closed")
	if err == nil {
		t.Error("changeIssueState without token expected error")
	}
}

func TestChangeIssueState_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	err := changeIssueState(app, "badrepo", "I1", "closed")
	if err == nil {
		t.Error("changeIssueState with bad repo expected error")
	}
}

func TestIssueList_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue list with API error expected error")
	}
}

func TestIssueView_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"view", "I1", "--repo", "owner/repo"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue view with API error expected error")
	}
}

func TestIssueCreate_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "owner/repo", "--title", "Bug"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue create with API error expected error")
	}
}

func TestIssueComment_APIError(t *testing.T) {
	app := testErrorApp()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"comment", "I1", "--repo", "owner/repo", "--body", "test"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue comment with API error expected error")
	}
}

func TestChangeIssueState_APIError(t *testing.T) {
	app := testErrorApp()
	err := changeIssueState(app, "owner/repo", "I1", "closed")
	if err == nil {
		t.Error("changeIssueState with API error expected error")
	}
}

func TestIssueStatus_AssignedError(t *testing.T) {
	app := testErrorApp()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err == nil {
		t.Error("issue status with API error expected error")
	}
}

func TestIssueList_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"list", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestIssueView_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"view", "I1", "--repo", "badrepo"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestIssueCreate_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"create", "--repo", "badrepo", "--title", "Bug"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestIssueStatus_CreatedError(t *testing.T) {
	// Create a server that returns OK for assigned but errors for created
	mux := http.NewServeMux()
	calls := 0
	mux.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			// First call (assigned) succeeds
			json.NewEncoder(w).Encode([]api.Issue{})
		} else {
			// Second call (created) errors
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Default()
	cfg.APIBase = srv.URL
	app := &App{
		Cfg:         cfg,
		Client:      api.New(srv.URL, "token"),
		ActiveToken: "token",
		Ctx:         context.Background(),
	}

	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for created issues API failure")
	}
}

func TestIssueComment_BadRepo(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()
	cmd := newIssueCmd(app)
	cmd.SetArgs([]string{"comment", "I1", "--repo", "badrepo", "--body", "test"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}
