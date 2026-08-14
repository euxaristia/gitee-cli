package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/euxaristia/gitee-cli/internal/api"
)

func TestSplitHeader(t *testing.T) {
	k, v, err := splitHeader("X-Request-Id: abc")
	if err != nil || k != "X-Request-Id" || v != "abc" {
		t.Fatalf("splitHeader() = %q, %q, %v", k, v, err)
	}
	if _, _, err := splitHeader("nocolon"); err == nil {
		t.Fatal("splitHeader(nocolon) expected error")
	}
	if _, _, err := splitHeader(":value"); err == nil {
		t.Fatal("splitHeader(:value) expected error")
	}
}

func TestAPICmd_CustomHeader(t *testing.T) {
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/test/endpoint", func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Request-Id")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	app := testAppNoToken()
	app.ActiveToken = "test-token"
	app.Client = api.New(srv.URL, "test-token")
	app.Cfg.APIBase = srv.URL

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-H", "X-Request-Id: abc-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("api with header error = %v", err)
	}
	if got != "abc-123" {
		t.Errorf("X-Request-Id = %q, want abc-123", got)
	}
}

func TestAPICmd_GET(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("api GET error = %v", err)
	}
}

func TestAPICmd_GET_WithFields(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-F", "key=value"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("api GET with fields error = %v", err)
	}
}

func TestAPICmd_POST(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-X", "POST", "-F", "name=test"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("api POST error = %v", err)
	}
}

func TestAPICmd_DELETE(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-X", "DELETE"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("api DELETE error = %v", err)
	}
}

func TestAPICmd_POST_NoToken(t *testing.T) {
	app := testAppNoToken()
	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-X", "POST"})
	if err := cmd.Execute(); err == nil {
		t.Error("api POST without token expected error")
	}
}

func TestAPICmd_POST_NoFields(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-X", "POST"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("api POST no fields error = %v", err)
	}
}

func TestAPICmd_InvalidField(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newAPICmd(app)
	cmd.SetArgs([]string{"/test/endpoint", "-F", "badfield"})
	if err := cmd.Execute(); err == nil {
		t.Error("api with invalid field expected error")
	}
}
