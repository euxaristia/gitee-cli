package cmd

import (
	"testing"
)

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
