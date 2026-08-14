package main

import (
	"fmt"
	"os"

	"github.com/euxaristia/gitee-cli/internal/cmd"
)

func main() {
	root := cmd.NewRootCmd()
	root.SetArgs(os.Args[1:])
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
