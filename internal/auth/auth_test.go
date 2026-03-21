package auth

import (
	"os"
	"testing"
)

func TestReadTokenFromTTY_NotTerminal(t *testing.T) {
	// When stdin is a pipe (not a terminal), term.ReadPassword returns an error.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	_, err = ReadTokenFromTTY()
	if err == nil {
		t.Error("ReadTokenFromTTY() expected error for non-terminal stdin, got nil")
	}
}
