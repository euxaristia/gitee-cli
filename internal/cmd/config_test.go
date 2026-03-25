package cmd

import (
	"testing"

	"github.com/euxaristia/gitee-cli/internal/config"
)

func TestGetConfigValue(t *testing.T) {
	cfg := config.Default()
	cfg.Editor = "vim"
	cfg.CurrentOrg = "myorg"
	cfg.GitFlags = []string{"--verbose", "--no-pager"}

	tests := []struct {
		key  string
		want string
	}{
		{"host", cfg.Host},
		{"api_base", cfg.APIBase},
		{"token", "(managed by keychain; use `gt auth token`)"},
		{"output", cfg.Output},
		{"editor", "vim"},
		{"current_org", "myorg"},
		{"git_protocol", cfg.GitProtocol},
		{"git_flags", "--verbose,--no-pager"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := getConfigValue(cfg, tt.key)
			if err != nil {
				t.Fatalf("getConfigValue(%q) error = %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("getConfigValue(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestGetConfigValue_Unknown(t *testing.T) {
	cfg := config.Default()
	_, err := getConfigValue(cfg, "nonexistent")
	if err == nil {
		t.Error("getConfigValue(nonexistent) expected error")
	}
}

func TestSetConfigValue(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{"host", "https://custom.gitee.com"},
		{"api_base", "https://custom.gitee.com/api/v5"},
		{"output", "json"},
		{"editor", "nano"},
		{"current_org", "myorg"},
		{"git_protocol", "ssh"},
		{"git_flags", "a,b,c"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			cfg := config.Default()
			err := setConfigValue(cfg, tt.key, tt.value)
			if err != nil {
				t.Fatalf("setConfigValue(%q, %q) error = %v", tt.key, tt.value, err)
			}
		})
	}
}

func TestSetConfigValue_GitFlagsEmpty(t *testing.T) {
	cfg := config.Default()
	cfg.GitFlags = []string{"--verbose"}
	if err := setConfigValue(cfg, "git_flags", ""); err != nil {
		t.Fatalf("setConfigValue(git_flags, '') error = %v", err)
	}
	if len(cfg.GitFlags) != 0 {
		t.Errorf("GitFlags = %v, want empty", cfg.GitFlags)
	}
}

func TestSetConfigValue_Token(t *testing.T) {
	cfg := config.Default()
	err := setConfigValue(cfg, "token", "value")
	if err == nil {
		t.Error("setConfigValue(token) expected error")
	}
}

func TestSetConfigValue_Unknown(t *testing.T) {
	cfg := config.Default()
	err := setConfigValue(cfg, "nonexistent", "value")
	if err == nil {
		t.Error("setConfigValue(nonexistent) expected error")
	}
}

func TestNewConfigCmd_List(t *testing.T) {
	_, app := testServer()
	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("config list error = %v", err)
	}
}

func TestNewConfigCmd_ListJSON(t *testing.T) {
	_, app := testServer()
	app.Cfg.Output = "json"
	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("config list json error = %v", err)
	}
}

func TestNewConfigCmd_Get(t *testing.T) {
	_, app := testServer()
	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"get", "host"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("config get error = %v", err)
	}
}

func TestNewConfigCmd_Get_Unknown(t *testing.T) {
	_, app := testServer()
	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"get", "nonexistent"})
	if err := cmd.Execute(); err == nil {
		t.Error("config get nonexistent expected error")
	}
}

func TestNewConfigCmd_Set(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"set", "editor", "nano"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("config set error = %v", err)
	}
	if app.Cfg.Editor != "nano" {
		t.Errorf("Editor = %q, want nano", app.Cfg.Editor)
	}
}

func TestNewConfigCmd_Set_Unknown(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"set", "nonexistent", "value"})
	if err := cmd.Execute(); err == nil {
		t.Error("config set nonexistent expected error")
	}
}

func TestNewConfigCmd_Unset(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	app.Cfg.Editor = "vim"
	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"unset", "editor"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("config unset error = %v", err)
	}
	if app.Cfg.Editor != "" {
		t.Errorf("Editor = %q, want empty", app.Cfg.Editor)
	}
}

func TestNewConfigCmd_Set_SaveError(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	// Make HOME invalid so config.Save fails
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"set", "editor", "nano"})
	if err := cmd.Execute(); err == nil {
		t.Error("config set with save error expected error")
	}
}

func TestNewConfigCmd_Unset_Unknown(t *testing.T) {
	srv, app := testServer()
	defer srv.Close()

	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"unset", "nonexistent"})
	if err := cmd.Execute(); err == nil {
		t.Error("config unset nonexistent expected error")
	}
}

func TestNewConfigCmd_Path(t *testing.T) {
	_, app := testServer()
	cmd := newConfigCmd(app)
	cmd.SetArgs([]string{"path"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("config path error = %v", err)
	}
}
