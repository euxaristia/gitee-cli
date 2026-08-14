package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Host != "https://gitee.com" {
		t.Errorf("Host = %q, want %q", cfg.Host, "https://gitee.com")
	}
	if cfg.APIBase != "https://gitee.com/api/v5" {
		t.Errorf("APIBase = %q, want %q", cfg.APIBase, "https://gitee.com/api/v5")
	}
	if cfg.Output != "table" {
		t.Errorf("Output = %q, want %q", cfg.Output, "table")
	}
	if cfg.Aliases == nil {
		t.Error("Aliases should not be nil")
	}
	if cfg.GitProtocol != "https" {
		t.Errorf("GitProtocol = %q, want %q", cfg.GitProtocol, "https")
	}
}

func TestConfigPath(t *testing.T) {
	p, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}
	if !filepath.IsAbs(p) {
		t.Errorf("ConfigPath() = %q, want absolute path", p)
	}
	if filepath.Base(p) != "config.yaml" {
		t.Errorf("ConfigPath() base = %q, want config.yaml", filepath.Base(p))
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	// Point HOME to a temp dir with no config file
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	// Should return defaults
	if cfg.Host != "https://gitee.com" {
		t.Errorf("Host = %q, want default", cfg.Host)
	}
	if cfg.Output != "table" {
		t.Errorf("Output = %q, want default", cfg.Output)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "gitee-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("host: https://custom.gitee.com\napi_base: https://custom.gitee.com/api/v5\noutput: json\neditor: vim\ncurrent_org: myorg\nuser: testuser\ngit_protocol: ssh\ngit_flags:\n  - --verbose\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "https://custom.gitee.com" {
		t.Errorf("Host = %q", cfg.Host)
	}
	if cfg.Output != "json" {
		t.Errorf("Output = %q", cfg.Output)
	}
	if cfg.Editor != "vim" {
		t.Errorf("Editor = %q", cfg.Editor)
	}
	if cfg.GitProtocol != "ssh" {
		t.Errorf("GitProtocol = %q", cfg.GitProtocol)
	}
	if len(cfg.GitFlags) != 1 || cfg.GitFlags[0] != "--verbose" {
		t.Errorf("GitFlags = %v", cfg.GitFlags)
	}
}

func TestLoad_EmptyFields(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "gitee-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write config with all empty fields to test default filling
	content := []byte("editor: nano\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != defaultHost {
		t.Errorf("Host = %q, want default", cfg.Host)
	}
	if cfg.APIBase != defaultAPIBase {
		t.Errorf("APIBase = %q, want default", cfg.APIBase)
	}
	if cfg.Output != "table" {
		t.Errorf("Output = %q, want table", cfg.Output)
	}
	if cfg.Aliases == nil {
		t.Error("Aliases should not be nil")
	}
	if cfg.GitProtocol != "https" {
		t.Errorf("GitProtocol = %q, want https", cfg.GitProtocol)
	}
	if cfg.GitFlags == nil {
		t.Error("GitFlags should not be nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "gitee-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":::invalid"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid YAML")
	}
}

func TestConfigPath_Error(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("UserConfigDir does not require HOME on Windows")
	}
	// Unsetting HOME and XDG_CONFIG_HOME causes UserConfigDir to fail
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := ConfigPath()
	if err == nil {
		t.Error("ConfigPath() expected error when HOME is unset")
	}
}

func TestLoad_ConfigPathError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("UserConfigDir does not require HOME on Windows")
	}
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error when ConfigPath fails")
	}
}

func TestLoad_ReadFileError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Create config.yaml as a directory to trigger a read error that isn't ErrNotExist
	dir := filepath.Join(tmp, "gitee-cli", "config.yaml")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error when config path is a directory")
	}
}

func TestSave_ConfigPathError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("UserConfigDir does not require HOME on Windows")
	}
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	err := Save(Default())
	if err == nil {
		t.Error("Save() expected error when ConfigPath fails")
	}
}

func TestSave_MkdirAllError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Create a file at the path where the directory should be, to make MkdirAll fail
	if err := os.WriteFile(filepath.Join(tmp, "gitee-cli"), []byte("block"), 0o444); err != nil {
		t.Fatal(err)
	}

	err := Save(Default())
	if err == nil {
		t.Error("Save() expected error when MkdirAll fails")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := Default()
	cfg.Editor = "code"
	cfg.CurrentOrg = "myorg"

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Editor != "code" {
		t.Errorf("Editor = %q, want code", loaded.Editor)
	}
	if loaded.CurrentOrg != "myorg" {
		t.Errorf("CurrentOrg = %q, want myorg", loaded.CurrentOrg)
	}
}
