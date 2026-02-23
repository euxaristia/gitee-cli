package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/config"
	"github.com/euxaristia/gitee-cli/internal/output"
)

func newConfigCmd(app *App) *cobra.Command {
	cfgCmd := &cobra.Command{Use: "config", Short: "Manage gitee CLI config"}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
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
				{"token", "(managed by keychain; use `gitee auth` commands)"},
			}
			output.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := getConfigValue(app.Cfg, args[0])
			if err != nil {
				return err
			}
			fmt.Println(v)
			return nil
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setConfigValue(app.Cfg, args[0], args[1]); err != nil {
				return err
			}
			return config.Save(app.Cfg)
		},
	}

	unsetCmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a config key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setConfigValue(app.Cfg, args[0], ""); err != nil {
				return err
			}
			return config.Save(app.Cfg)
		},
	}

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.ConfigPath()
			if err != nil {
				return err
			}
			fmt.Println(p)
			return nil
		},
	}

	cfgCmd.AddCommand(listCmd, getCmd, setCmd, unsetCmd, pathCmd)
	return cfgCmd
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "host":
		return cfg.Host, nil
	case "api_base":
		return cfg.APIBase, nil
	case "token":
		return "(managed by keychain; use `gitee auth token`)", nil
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
		return fmt.Errorf("token is managed by keychain; use `gitee auth login/logout`")
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
