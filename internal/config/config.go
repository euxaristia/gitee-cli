package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	defaultHost    = "https://gitee.com"
	defaultAPIBase = "https://gitee.com/api/v5"
)

type Config struct {
	Host       string            `yaml:"host"`
	APIBase    string            `yaml:"api_base"`
	Token      string            `yaml:"token"`
	Output     string            `yaml:"output"`
	Editor     string            `yaml:"editor"`
	Aliases    map[string]string `yaml:"aliases"`
	CurrentOrg string            `yaml:"current_org"`
}

func Default() *Config {
	return &Config{
		Host:    defaultHost,
		APIBase: defaultAPIBase,
		Output:  "table",
		Aliases: map[string]string{},
	}
}

func ConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gitee-cli", "config.yaml"), nil
}

func Load() (*Config, error) {
	cfg := Default()
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.APIBase == "" {
		cfg.APIBase = defaultAPIBase
	}
	if cfg.Output == "" {
		cfg.Output = "table"
	}
	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
