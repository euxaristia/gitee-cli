package cmd

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var asciiArt = []string{
	" dP\"\"b8 88 888888 888888 888888           dP\"\"b8 88     88 ",
	"dP   `\" 88   88   88__   88__   ________ dP   `\" 88     88 ",
	"Yb  \"88 88   88   88\"\"   88\"\"   \"\"\"\"\"\"\" Yb      88  .o 88 ",
	" YboodP 88   88   888888 888888           YboodP 88ood8 88 ",
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

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			printVersionBanner(cmd.OutOrStdout())
		},
	}
}
