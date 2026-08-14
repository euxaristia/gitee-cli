package cmd

import (
	"fmt"
	"strings"

	"github.com/euxaristia/gitee-cli/internal/config"
	"github.com/euxaristia/gitee-cli/internal/output"
)

func newConfigCmd(app *App) *Command {
	return &Command{
		Use:   "config",
		Short: "Manage gitee CLI config",
		run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("config requires a subcommand")
			}
			switch args[0] {
			case "list":
				return runConfigList(app)
			case "get":
				return runConfigGet(app, args[1:])
			case "set":
				return runConfigSet(app, args[1:])
			case "unset":
				return runConfigUnset(app, args[1:])
			case "path":
				return runConfigPath()
			default:
				return fmt.Errorf("unknown config command %q", args[0])
			}
		},
	}
}

func runConfigList(app *App) error {
	if app.Cfg.Output == string(output.FormatJSON) {
		return output.PrintJSON(app.Cfg)
	}
	rows := [][]string{
		{"host", app.Cfg.Host},
		{"api_base", app.Cfg.APIBase},
		{"output", app.Cfg.Output},
		{"editor", app.Cfg.Editor},
		{"git_protocol", app.Cfg.GitProtocol},
		{"git_flags", strings.Join(app.Cfg.GitFlags, ",")},
		{"token", "(managed by keychain; use `gt auth` commands)"},
	}
	output.PrintTable([]string{"KEY", "VALUE"}, rows)
	return nil
}

func runConfigGet(app *App, args []string) error {
	if err := exactArgs(args, 1); err != nil {
		return err
	}
	v, err := getConfigValue(app.Cfg, args[0])
	if err != nil {
		return err
	}
	fmt.Println(v)
	return nil
}

func runConfigSet(app *App, args []string) error {
	if err := exactArgs(args, 2); err != nil {
		return err
	}
	if err := setConfigValue(app.Cfg, args[0], args[1]); err != nil {
		return err
	}
	return config.Save(app.Cfg)
}

func runConfigUnset(app *App, args []string) error {
	if err := exactArgs(args, 1); err != nil {
		return err
	}
	if err := setConfigValue(app.Cfg, args[0], ""); err != nil {
		return err
	}
	return config.Save(app.Cfg)
}

func runConfigPath() error {
	p, err := config.ConfigPath()
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "host":
		return cfg.Host, nil
	case "api_base":
		return cfg.APIBase, nil
	case "token":
		return "(managed by keychain; use `gt auth token`)", nil
	case "output":
		return cfg.Output, nil
	case "editor":
		return cfg.Editor, nil
	case "current_org":
		return cfg.CurrentOrg, nil
	case "git_protocol":
		return cfg.GitProtocol, nil
	case "git_flags":
		return strings.Join(cfg.GitFlags, ","), nil
	default:
		return "", fmt.Errorf("unknown key: %s", key)
	}
}

func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "host":
		cfg.Host = value
	case "api_base":
		cfg.APIBase = value
	case "token":
		return fmt.Errorf("token is managed by keychain; use `gt auth login/logout`")
	case "output":
		cfg.Output = value
	case "editor":
		cfg.Editor = value
	case "current_org":
		cfg.CurrentOrg = value
	case "git_protocol":
		cfg.GitProtocol = value
	case "git_flags":
		if value == "" {
			cfg.GitFlags = []string{}
		} else {
			cfg.GitFlags = strings.Split(value, ",")
		}
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	return nil
}
