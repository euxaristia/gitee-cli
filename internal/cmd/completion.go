package cmd

import (
	"fmt"
	"io"
)

func newCompletionCmd() *Command {
	return &Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		run: func(c *Command, args []string) error {
			if err := exactArgs(args, 1); err != nil {
				return err
			}
			return writeCompletion(c.OutOrStdout(), args[0])
		},
	}
}

func writeCompletion(w io.Writer, shell string) error {
	switch shell {
	case "bash":
		fmt.Fprint(w, bashCompletion)
	case "zsh":
		fmt.Fprint(w, zshCompletion)
	case "fish":
		fmt.Fprint(w, fishCompletion)
	case "powershell":
		fmt.Fprint(w, powershellCompletion)
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
	return nil
}

const completionWords = "api auth commit completion config git issue pr pull push release repo status version"

const bashCompletion = `#!/usr/bin/env bash
_gt() {
  COMPREPLY=($(compgen -W "` + completionWords + `" -- "${COMP_WORDS[COMP_CWORD]}"))
}
complete -F _gt gt
`

const zshCompletion = `#compdef gt
_arguments '1: :(` + completionWords + `)'
`

const fishCompletion = `complete -c gt -f -a "` + completionWords + `"
`

const powershellCompletion = `Register-ArgumentCompleter -Native -CommandName gt -ScriptBlock {
  param($wordToComplete)
  '` + completionWords + `'.Split(' ') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object { $_ }
}
`
