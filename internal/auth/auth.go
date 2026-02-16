package auth

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func ReadTokenFromTTY() (string, error) {
	fmt.Fprint(os.Stdout, "Gitee personal access token: ")
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stdout)
	return strings.TrimSpace(string(b)), nil
}
