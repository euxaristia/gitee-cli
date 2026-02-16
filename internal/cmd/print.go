package cmd

import (
	"fmt"

	"github.com/euxaristia/gitee-cli/internal/output"
)

func printAny(format string, tableHeaders []string, tableRows [][]string, payload any) error {
	switch format {
	case "json":
		return output.PrintJSON(payload)
	case "table", "":
		output.PrintTable(tableHeaders, tableRows)
		return nil
	default:
		return fmt.Errorf("unknown output format %q", format)
	}
}
