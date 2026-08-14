package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "gt" {
		t.Errorf("root cmd Use = %q, want gt", cmd.Use)
	}
	if !cmd.SilenceUsage {
		t.Error("root cmd should silence usage")
	}
}

func TestNewRootCmd_Help(t *testing.T) {
	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()
}

func TestNewRootCmd_VersionSubcmd(t *testing.T) {
	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("root version error = %v", err)
	}
}

func TestNewRootCmd_WithOutputFlag(t *testing.T) {
	tmp := t.TempDir()
	useConfigDir(t, tmp)
	t.Setenv("GITEE_TOKEN", "test-token")

	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"-o", "json", "version"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("root with output flag error = %v", err)
	}
}

func TestNewRootCmd_PersistentPreRunE_LoadsConfig(t *testing.T) {
	tmp := t.TempDir()
	useConfigDir(t, tmp)
	t.Setenv("GITEE_TOKEN", "test-token")

	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"config", "path"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("root PersistentPreRunE error = %v", err)
	}
}

func TestNewRootCmd_PersistentPreRunE_NoEnvToken(t *testing.T) {
	tmp := t.TempDir()
	useConfigDir(t, tmp)
	t.Setenv("GITEE_TOKEN", "")

	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"config", "path"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("root PersistentPreRunE no env token error = %v", err)
	}
}

func TestNewRootCmd_RunE_NoVersion(t *testing.T) {
	tmp := t.TempDir()
	useConfigDir(t, tmp)
	t.Setenv("GITEE_TOKEN", "test")

	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Errorf("root RunE error = %v", err)
	}
}

func TestNewRootCmd_ConfigLoadError(t *testing.T) {
	tmp := t.TempDir()
	useConfigDir(t, tmp)
	t.Setenv("GITEE_TOKEN", "test")

	// Create config.yaml as a directory to cause Load error
	dir := filepath.Join(tmp, "gitee-cli", "config.yaml")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"config", "path"})
	err := cmd.Execute()
	if err == nil {
		t.Error("root with broken config expected error")
	}
}

func TestNewRootCmd_LegacyToken(t *testing.T) {
	tmp := t.TempDir()
	useConfigDir(t, tmp)
	t.Setenv("GITEE_TOKEN", "")

	// Write config with a legacy token
	dir := filepath.Join(tmp, "gitee-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("token: legacy-token-123\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"config", "path"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("root legacy token error = %v", err)
	}
}
