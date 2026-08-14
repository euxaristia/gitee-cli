package cmd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/euxaristia/gitee-cli/internal/api"
	"github.com/euxaristia/gitee-cli/internal/config"
)

func useConfigDir(t *testing.T, dir string) {
	t.Helper()
	orig := config.UserConfigDir
	config.UserConfigDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { config.UserConfigDir = orig })
}

func useConfigDirError(t *testing.T) {
	t.Helper()
	orig := config.UserConfigDir
	config.UserConfigDir = func() (string, error) { return "", errors.New("no config dir") }
	t.Cleanup(func() { config.UserConfigDir = orig })
}

// testServer creates a mock Gitee API server and returns the server and an App configured to use it.
func testServer() (*httptest.Server, *App) {
	mux := http.NewServeMux()

	// User
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(api.User{ID: 1, Login: "testuser", Name: "Test User"})
	})

	// Repos
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(api.Repo{ID: 1, FullName: "testuser/newrepo", HTMLURL: "https://gitee.com/testuser/newrepo"})
			return
		}
		json.NewEncoder(w).Encode([]api.Repo{
			{ID: 1, FullName: "testuser/repo1", HTMLURL: "https://gitee.com/testuser/repo1", DefaultBr: "main"},
		})
	})

	mux.HandleFunc("/orgs/myorg/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(api.Repo{ID: 2, FullName: "myorg/newrepo", HTMLURL: "https://gitee.com/myorg/newrepo"})
			return
		}
		json.NewEncoder(w).Encode([]api.Repo{
			{ID: 2, FullName: "myorg/repo1", HTMLURL: "https://gitee.com/myorg/repo1", DefaultBr: "main"},
		})
	})

	mux.HandleFunc("/repos/owner/repo", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(api.Repo{ID: 1, FullName: "owner/repo", HTMLURL: "https://gitee.com/owner/repo", DefaultBr: "main"})
	})

	// Issues
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]api.Issue{
			{ID: 1, Number: "I1", Title: "Bug", State: "open", User: api.User{Login: "testuser"}},
		})
	})

	mux.HandleFunc("/repos/owner/repo/issues/I1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(api.Issue{ID: 1, Number: "I1", Title: "Bug", State: "open", User: api.User{Login: "testuser"}, HTMLURL: "https://gitee.com/owner/repo/issues/I1"})
	})

	mux.HandleFunc("/repos/owner/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(api.Issue{ID: 2, Number: "I2", Title: "New Issue", HTMLURL: "https://gitee.com/owner/repo/issues/I2"})
	})

	mux.HandleFunc("/repos/owner/issues/I1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(api.Issue{ID: 1, Number: "I1", Title: "Bug", State: "closed", HTMLURL: "https://gitee.com/owner/repo/issues/I1"})
	})

	mux.HandleFunc("/repos/owner/repo/issues/I1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "1"})
	})

	mux.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if filter == "assigned" {
			json.NewEncoder(w).Encode([]api.Issue{
				{Number: "I1", Title: "Assigned issue"},
			})
		} else {
			json.NewEncoder(w).Encode([]api.Issue{})
		}
	})

	// PRs
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(api.PullRequest{ID: 1, Number: 1, Title: "New PR", HTMLURL: "https://gitee.com/owner/repo/pulls/1"})
			return
		}
		json.NewEncoder(w).Encode([]api.PullRequest{
			{ID: 1, Number: 1, Title: "PR 1", State: "open", User: api.User{Login: "testuser"}},
		})
	})

	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			json.NewEncoder(w).Encode(api.PullRequest{ID: 1, Number: 1, Title: "PR 1", State: "closed"})
			return
		}
		json.NewEncoder(w).Encode(api.PullRequest{ID: 1, Number: 1, Title: "PR 1", State: "open", User: api.User{Login: "testuser"}, HTMLURL: "https://gitee.com/owner/repo/pulls/1"})
	})

	mux.HandleFunc("/repos/owner/repo/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/repos/owner/repo/pulls/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "1"})
	})

	// Releases
	mux.HandleFunc("/repos/owner/repo/releases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(api.Release{ID: 1, TagName: "v1.0.0", Name: "Release 1.0", HTMLURL: "https://gitee.com/owner/repo/releases/v1.0.0"})
			return
		}
		json.NewEncoder(w).Encode([]api.Release{
			{ID: 1, TagName: "v1.0.0", Name: "Release 1.0", HTMLURL: "https://gitee.com/owner/repo/releases/v1.0.0"},
		})
	})

	mux.HandleFunc("/repos/owner/repo/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(api.Release{ID: 1, TagName: "v1.0.0", Name: "Release 1.0", HTMLURL: "https://gitee.com/owner/repo/releases/v1.0.0"})
	})

	mux.HandleFunc("/repos/owner/repo/releases/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(api.Release{ID: 1, TagName: "v1.0.0"})
	})

	// Catch-all for raw API command tests
	mux.HandleFunc("/test/endpoint", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	server := httptest.NewServer(mux)

	cfg := config.Default()
	cfg.APIBase = server.URL

	app := &App{
		Cfg:         cfg,
		Client:      api.New(server.URL, "test-token"),
		ActiveToken: "test-token",
		Ctx:         context.Background(),
	}

	return server, app
}

func testAppNoToken() *App {
	cfg := config.Default()
	return &App{
		Cfg:         cfg,
		Client:      api.New(cfg.APIBase, ""),
		ActiveToken: "",
		Ctx:         context.Background(),
	}
}

// testErrorApp returns an App pointing at an unreachable server (to trigger API errors).
func testErrorApp() *App {
	cfg := config.Default()
	cfg.APIBase = "http://127.0.0.1:1"
	return &App{
		Cfg:         cfg,
		Client:      api.New("http://127.0.0.1:1", "test-token"),
		ActiveToken: "test-token",
		Ctx:         context.Background(),
	}
}
