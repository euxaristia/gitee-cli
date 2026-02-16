package util

import (
	"fmt"
	"strings"
)

func SplitRepo(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo %q, expected owner/name", repo)
	}
	return parts[0], parts[1], nil
}

func KeyValue(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid key=value pair: %s", s)
	}
	return parts[0], parts[1], nil
}
