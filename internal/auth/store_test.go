package auth

import (
	"testing"

	keyring "github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func TestSaveAndLoadToken(t *testing.T) {
	token := "test-token-12345"
	if err := SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	got, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if got != token {
		t.Errorf("LoadToken() = %q, want %q", got, token)
	}
}

func TestLoadToken_NotFound(t *testing.T) {
	keyring.MockInit() // reset mock
	got, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if got != "" {
		t.Errorf("LoadToken() = %q, want empty string", got)
	}
}

func TestDeleteToken(t *testing.T) {
	if err := SaveToken("to-delete"); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}
	got, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() after delete error = %v", err)
	}
	if got != "" {
		t.Errorf("LoadToken() after delete = %q, want empty", got)
	}
}

func TestDeleteToken_NotFound(t *testing.T) {
	keyring.MockInit() // reset mock
	if err := DeleteToken(); err != nil {
		t.Errorf("DeleteToken() on missing token error = %v", err)
	}
}
