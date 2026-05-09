package util

import (
	"fmt"
	"os/exec"
	"strings"
)

// CurrentRepo returns the owner/name of the repository for the current directory.
func CurrentRepo() (string, error) {
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote url: %w", err)
	}
	urlStr := strings.TrimSpace(string(out))
	urlStr = strings.TrimSuffix(urlStr, ".git")
	urlStr = strings.TrimSuffix(urlStr, "/")

	// git@gitee.com:owner/repo
	if strings.HasPrefix(urlStr, "git@") {
		parts := strings.Split(urlStr, ":")
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	// https://gitee.com/owner/repo
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		parts := strings.Split(urlStr, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-2] + "/" + parts[len(parts)-1], nil
		}
	}

	return "", fmt.Errorf("could not parse remote URL: %s", urlStr)
}

// CurrentBranch returns the current git branch name.
func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
