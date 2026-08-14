package cmd

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var version = "0.1.0"

var asciiArt = []string{
	`  _______  __  .___________. _______  _______              ______  __       __  `,
	` /  _____||  | |           ||   ____||   ____|            /      ||  |     |  | `,
	`|  |  __  |  | ` + "`" + `---|  |----` + "`" + `|  |__   |  |__    ______    |  ,----'|  |     |  | `,
	`|  | |_ | |  |     |  |     |   __|  |   __|  |______|   |  |     |  |     |  | `,
	`|  |__| | |  |     |  |     |  |____ |  |____            |  ` + "`" + `----.|  ` + "`" + `----.|  | `,
	` \______| |__|     |__|     |_______||_______|            \______||_______||__| `,
}

func printVersionBanner(w io.Writer) {
	lines := append([]string{}, asciiArt...)
	lines = append(lines, "", fmt.Sprintf("Version: %s", version), "© 2026 euxaristia")

	width := 0
	for _, l := range lines {
		lineWidth := utf8.RuneCountInString(l)
		if lineWidth > width {
			width = lineWidth
		}
	}

	fmt.Fprintf(w, "+%s+\n", strings.Repeat("-", width+2))
	for _, l := range lines {
		padding := strings.Repeat(" ", width-utf8.RuneCountInString(l))
		fmt.Fprintf(w, "| %s%s |\n", l, padding)
	}
	fmt.Fprintf(w, "+%s+\n", strings.Repeat("-", width+2))
}

func newVersionCmd() *Command {
	return &Command{
		Use:   "version",
		Short: "Print version",
		run: func(c *Command, args []string) error {
			printVersionBanner(c.OutOrStdout())
			return nil
		},
	}
}
