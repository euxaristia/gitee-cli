package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setup(t *testing.T) (*httptest.Server, *Client) {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(User{ID: 1, Login: "testuser", Name: "Test User"})
	})

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(Repo{ID: 1, FullName: "testuser/repo"})
			return
		}
		json.NewEncoder(w).Encode([]Repo{{ID: 1, FullName: "testuser/repo"}})
	})

	mux.HandleFunc("/orgs/myorg/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(Repo{ID: 2, FullName: "myorg/repo"})
			return
		}
		json.NewEncoder(w).Encode([]Repo{{ID: 2, FullName: "myorg/repo"}})
	})

	mux.HandleFunc("/repos/owner/repo", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Repo{ID: 1, FullName: "owner/repo"})
	})

	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Issue{{ID: 1, Number: "I1", Title: "Bug"}})
	})

	mux.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Issue{{ID: 1, Number: "I1", Title: "My Issue"}})
	})

	mux.HandleFunc("/repos/owner/repo/issues/I1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Issue{ID: 1, Number: "I1", Title: "Bug"})
	})

	mux.HandleFunc("/repos/owner/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Issue{ID: 2, Number: "I2", Title: "New"})
	})

	mux.HandleFunc("/repos/owner/issues/I1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Issue{ID: 1, Number: "I1", State: "closed"})
	})

	mux.HandleFunc("/repos/owner/repo/issues/I1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(PullRequest{ID: 1, Number: 1, Title: "PR"})
			return
		}
		json.NewEncoder(w).Encode([]PullRequest{{ID: 1, Number: 1, Title: "PR"}})
	})

	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			json.NewEncoder(w).Encode(PullRequest{ID: 1, Number: 1, State: "closed"})
			return
		}
		json.NewEncoder(w).Encode(PullRequest{ID: 1, Number: 1, Title: "PR"})
	})

	mux.HandleFunc("/repos/owner/repo/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/repos/owner/repo/pulls/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	mux.HandleFunc("/repos/owner/repo/releases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(Release{ID: 1, TagName: "v1.0"})
			return
		}
		json.NewEncoder(w).Encode([]Release{{ID: 1, TagName: "v1.0"}})
	})

	mux.HandleFunc("/repos/owner/repo/releases/tags/v1.0", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Release{ID: 1, TagName: "v1.0"})
	})

	mux.HandleFunc("/repos/owner/repo/releases/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	})

	srv := httptest.NewServer(mux)
	client := New(srv.URL, "test-token")
	return srv, client
}

func TestNew(t *testing.T) {
	c := New("https://gitee.com/api/v5/", "mytoken")
	if c.baseURL != "https://gitee.com/api/v5" {
		t.Errorf("baseURL = %q, want trailing slash stripped", c.baseURL)
	}
	if c.token != "mytoken" {
		t.Errorf("token = %q", c.token)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestAPIError_Error(t *testing.T) {
	e := &APIError{StatusCode: 404, Body: "not found"}
	got := e.Error()
	if !strings.Contains(got, "404") || !strings.Contains(got, "not found") {
		t.Errorf("Error() = %q", got)
	}
}

func TestRequest_APIError(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	var out map[string]string
	err := client.Request(context.Background(), "GET", "/error", nil, nil, &out)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestRequest_WithQuery(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	var user User
	err := client.Request(context.Background(), "GET", "/user", map[string]string{"key": "val", "empty": ""}, nil, &user)
	if err != nil {
		t.Fatalf("Request error = %v", err)
	}
	if user.Login != "testuser" {
		t.Errorf("Login = %q", user.Login)
	}
}

func TestRequest_WithBody(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	body := map[string]string{"name": "repo"}
	var repo Repo
	err := client.Request(context.Background(), "POST", "/user/repos", nil, body, &repo)
	if err != nil {
		t.Fatalf("Request error = %v", err)
	}
}

func TestRequest_NilOut(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	err := client.Request(context.Background(), "DELETE", "/repos/owner/repo/releases/1", nil, nil, nil)
	if err != nil {
		t.Fatalf("Request error = %v", err)
	}
}

func TestRequest_NoToken(t *testing.T) {
	srv, _ := setup(t)
	defer srv.Close()

	client := New(srv.URL, "")
	var user User
	err := client.Request(context.Background(), "GET", "/user", nil, nil, &user)
	if err != nil {
		t.Fatalf("Request without token error = %v", err)
	}
}

func TestCurrentUser(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("CurrentUser error = %v", err)
	}
	if user.Login != "testuser" {
		t.Errorf("Login = %q", user.Login)
	}
}

func TestListRepos(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	repos, err := client.ListRepos(context.Background(), "", "all", 1, 30)
	if err != nil {
		t.Fatalf("ListRepos error = %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("len(repos) = %d", len(repos))
	}
}

func TestListRepos_WithOrg(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	repos, err := client.ListRepos(context.Background(), "myorg", "all", 1, 30)
	if err != nil {
		t.Fatalf("ListRepos with org error = %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("len(repos) = %d", len(repos))
	}
	if repos[0].FullName != "myorg/repo" {
		t.Errorf("FullName = %q", repos[0].FullName)
	}
}

func TestGetRepo(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	repo, err := client.GetRepo(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("GetRepo error = %v", err)
	}
	if repo.FullName != "owner/repo" {
		t.Errorf("FullName = %q", repo.FullName)
	}
}

func TestCreateRepo(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	repo, err := client.CreateRepo(context.Background(), "newrepo", "desc", "", false)
	if err != nil {
		t.Fatalf("CreateRepo error = %v", err)
	}
	if repo.FullName != "testuser/repo" {
		t.Errorf("FullName = %q", repo.FullName)
	}
}

func TestCreateRepo_WithOrg(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	repo, err := client.CreateRepo(context.Background(), "newrepo", "desc", "myorg", false)
	if err != nil {
		t.Fatalf("CreateRepo with org error = %v", err)
	}
	if repo.FullName != "myorg/repo" {
		t.Errorf("FullName = %q", repo.FullName)
	}
}

func TestListIssues(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	issues, err := client.ListIssues(context.Background(), "owner", "repo", "open", 1, 30)
	if err != nil {
		t.Fatalf("ListIssues error = %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("len(issues) = %d", len(issues))
	}
}

func TestListAllIssues(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	issues, err := client.ListAllIssues(context.Background(), "assigned", "open", 1, 10)
	if err != nil {
		t.Fatalf("ListAllIssues error = %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("len(issues) = %d", len(issues))
	}
}

func TestGetIssue(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	issue, err := client.GetIssue(context.Background(), "owner", "repo", "I1")
	if err != nil {
		t.Fatalf("GetIssue error = %v", err)
	}
	if issue.Number != "I1" {
		t.Errorf("Number = %q", issue.Number)
	}
}

func TestCreateIssue(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	issue, err := client.CreateIssue(context.Background(), "owner", "repo", "title", "body")
	if err != nil {
		t.Fatalf("CreateIssue error = %v", err)
	}
	if issue.Number != "I2" {
		t.Errorf("Number = %q", issue.Number)
	}
}

func TestUpdateIssueState(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	issue, err := client.UpdateIssueState(context.Background(), "owner", "repo", "I1", "closed")
	if err != nil {
		t.Fatalf("UpdateIssueState error = %v", err)
	}
	if issue.State != "closed" {
		t.Errorf("State = %q", issue.State)
	}
}

func TestCreateIssueComment(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	err := client.CreateIssueComment(context.Background(), "owner", "repo", "I1", "comment")
	if err != nil {
		t.Fatalf("CreateIssueComment error = %v", err)
	}
}

func TestListPRs(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	prs, err := client.ListPRs(context.Background(), "owner", "repo", "open", "", 1, 30)
	if err != nil {
		t.Fatalf("ListPRs error = %v", err)
	}
	if len(prs) != 1 {
		t.Errorf("len(prs) = %d", len(prs))
	}
}

func TestGetPR(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	pr, err := client.GetPR(context.Background(), "owner", "repo", 1)
	if err != nil {
		t.Fatalf("GetPR error = %v", err)
	}
	if pr.Number != 1 {
		t.Errorf("Number = %d", pr.Number)
	}
}

func TestCreatePR(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	pr, err := client.CreatePR(context.Background(), "owner", "repo", "title", "head", "base", "body")
	if err != nil {
		t.Fatalf("CreatePR error = %v", err)
	}
	if pr.Title != "PR" {
		t.Errorf("Title = %q", pr.Title)
	}
}

func TestMergePR(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	err := client.MergePR(context.Background(), "owner", "repo", 1, "merge msg")
	if err != nil {
		t.Fatalf("MergePR error = %v", err)
	}
}

func TestClosePR(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	pr, err := client.ClosePR(context.Background(), "owner", "repo", 1)
	if err != nil {
		t.Fatalf("ClosePR error = %v", err)
	}
	if pr.State != "closed" {
		t.Errorf("State = %q", pr.State)
	}
}

func TestCreatePRComment(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	err := client.CreatePRComment(context.Background(), "owner", "repo", 1, "nice")
	if err != nil {
		t.Fatalf("CreatePRComment error = %v", err)
	}
}

func TestListReleases(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	rels, err := client.ListReleases(context.Background(), "owner", "repo", 1, 30)
	if err != nil {
		t.Fatalf("ListReleases error = %v", err)
	}
	if len(rels) != 1 {
		t.Errorf("len(releases) = %d", len(rels))
	}
}

func TestGetReleaseByTag(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	rel, err := client.GetReleaseByTag(context.Background(), "owner", "repo", "v1.0")
	if err != nil {
		t.Fatalf("GetReleaseByTag error = %v", err)
	}
	if rel.TagName != "v1.0" {
		t.Errorf("TagName = %q", rel.TagName)
	}
}

func TestCreateRelease(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	rel, err := client.CreateRelease(context.Background(), "owner", "repo", "v1.0", "Release", "notes", "main", false)
	if err != nil {
		t.Fatalf("CreateRelease error = %v", err)
	}
	if rel.TagName != "v1.0" {
		t.Errorf("TagName = %q", rel.TagName)
	}
}

func TestDeleteRelease(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	err := client.DeleteRelease(context.Background(), "owner", "repo", 1)
	if err != nil {
		t.Fatalf("DeleteRelease error = %v", err)
	}
}

func TestIsTransientNetErr(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{fmt.Errorf("tls handshake timeout"), true},
		{fmt.Errorf("connection reset by peer"), true},
		{fmt.Errorf("unexpected eof"), true},
		{fmt.Errorf("normal error"), false},
	}
	for _, tt := range tests {
		if got := isTransientNetErr(tt.err); got != tt.want {
			t.Errorf("isTransientNetErr(%q) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestRequest_ConnectionRefused(t *testing.T) {
	// Test with unreachable server
	client := New("http://127.0.0.1:1", "token")
	var user User
	err := client.Request(context.Background(), "GET", "/user", nil, nil, &user)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestCurrentUser_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.CurrentUser(context.Background())
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetRepo_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.GetRepo(context.Background(), "o", "r")
	if err == nil {
		t.Error("expected error")
	}
}

func TestListIssues_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.ListIssues(context.Background(), "o", "r", "open", 1, 10)
	if err == nil {
		t.Error("expected error")
	}
}

func TestListAllIssues_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.ListAllIssues(context.Background(), "assigned", "open", 1, 10)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetIssue_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.GetIssue(context.Background(), "o", "r", "1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCreateIssue_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.CreateIssue(context.Background(), "o", "r", "t", "b")
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpdateIssueState_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.UpdateIssueState(context.Background(), "o", "r", "1", "closed")
	if err == nil {
		t.Error("expected error")
	}
}

func TestListPRs_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.ListPRs(context.Background(), "o", "r", "open", "", 1, 10)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetPR_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.GetPR(context.Background(), "o", "r", 1)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCreatePR_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.CreatePR(context.Background(), "o", "r", "t", "h", "b", "bd")
	if err == nil {
		t.Error("expected error")
	}
}

func TestClosePR_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.ClosePR(context.Background(), "o", "r", 1)
	if err == nil {
		t.Error("expected error")
	}
}

func TestListReleases_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.ListReleases(context.Background(), "o", "r", 1, 10)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetReleaseByTag_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.GetReleaseByTag(context.Background(), "o", "r", "v1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCreateRelease_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.CreateRelease(context.Background(), "o", "r", "t", "n", "b", "m", false)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCreateRepo_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.CreateRepo(context.Background(), "n", "d", "", false)
	if err == nil {
		t.Error("expected error")
	}
}

func TestListRepos_Error(t *testing.T) {
	client := New("http://127.0.0.1:1", "token")
	_, err := client.ListRepos(context.Background(), "", "all", 1, 10)
	if err == nil {
		t.Error("expected error")
	}
}

// netTimeoutErr implements net.Error for testing isTransientNetErr.
type netTimeoutErr struct{}

func (e *netTimeoutErr) Error() string   { return "timeout" }
func (e *netTimeoutErr) Timeout() bool   { return true }
func (e *netTimeoutErr) Temporary() bool { return false }

func TestIsTransientNetErr_NetTimeout(t *testing.T) {
	if !isTransientNetErr(&netTimeoutErr{}) {
		t.Error("expected true for net.Error with Timeout()")
	}
}

func TestDoWithRetry_POST_NoRetry(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "error", 500)
	}))
	defer srv.Close()

	client := New(srv.URL, "token")
	var out any
	_ = client.Request(context.Background(), "POST", "/test", nil, map[string]string{"k": "v"}, &out)
	if calls != 1 {
		t.Errorf("POST made %d calls, want 1", calls)
	}
}

func TestRequest_BadURL(t *testing.T) {
	client := &Client{
		baseURL:    "://invalid",
		token:      "token",
		httpClient: http.DefaultClient,
	}
	err := client.Request(context.Background(), "GET", "/test", nil, nil, nil)
	if err == nil {
		t.Error("expected error for bad URL")
	}
}

func TestRequest_BadBody(t *testing.T) {
	srv, client := setup(t)
	defer srv.Close()

	// Channels can't be JSON marshaled
	err := client.Request(context.Background(), "POST", "/test", nil, make(chan int), nil)
	if err == nil {
		t.Error("expected error for unmarshalable body")
	}
}

func TestNewRequest_BadURL(t *testing.T) {
	client := New("http://localhost", "token")
	_, err := client.newRequest(context.Background(), "GET", "://invalid\x00url", nil)
	if err == nil {
		t.Error("expected error for bad URL in newRequest")
	}
}
