package util

import (
	"testing"
)

func TestSplitRepo(t *testing.T) {
	tests := []struct {
		repo    string
		owner   string
		name    string
		wantErr bool
	}{
		{"owner/repo", "owner", "repo", false},
		{"owner/", "", "", true},
		{"/repo", "", "", true},
		{"owner/repo/extra", "", "", true},
		{"justrepo", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range tests {
		owner, name, err := SplitRepo(tt.repo)
		if (err != nil) != tt.wantErr {
			t.Errorf("SplitRepo(%q) error = %v, wantErr %v", tt.repo, err, tt.wantErr)
			return
		}
		if owner != tt.owner || name != tt.name {
			t.Errorf("SplitRepo(%q) = (%q, %q), want (%q, %q)", tt.repo, owner, name, tt.owner, tt.name)
		}
	}
}

func TestKeyValue(t *testing.T) {
	tests := []struct {
		s       string
		key     string
		val     string
		wantErr bool
	}{
		{"key=value", "key", "value", false},
		{"key=", "key", "", false},
		{"=value", "", "", true},
		{"key", "", "", true},
		{"", "", "", true},
		{"key=value=more", "key", "value=more", false},
	}

	for _, tt := range tests {
		k, v, err := KeyValue(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("KeyValue(%q) error = %v, wantErr %v", tt.s, err, tt.wantErr)
			return
		}
		if k != tt.key || v != tt.val {
			t.Errorf("KeyValue(%q) = (%q, %q), want (%q, %q)", tt.s, k, v, tt.key, tt.val)
		}
	}
}
