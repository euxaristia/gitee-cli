package util

import (
	"fmt"
	"os/exec"
	"strings"
)

// upstreamRemote returns the remote name for the current branch's upstream,
// falling back to "origin".
func upstreamRemote() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "@{upstream}").Output()
	if err != nil {
		return "origin"
	}
	s := strings.TrimSpace(string(out))
	if i := strings.Index(s, "/"); i != -1 {
		return s[:i]
	}
	return "origin"
}

// CurrentRepo returns the owner/name of the repository for the current directory.
func CurrentRepo() (string, error) {
	remote := upstreamRemote()
	out, err := exec.Command("git", "config", "--get", "remote."+remote+".url").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote url: %w", err)
	}
	urlStr := strings.TrimSpace(string(out))
	urlStr = strings.TrimSuffix(urlStr, ".git")
	urlStr = strings.TrimSuffix(urlStr, "/")

	// Strip protocol and user prefix for any URL scheme (git@, ssh://, http(s)://)
	if i := strings.Index(urlStr, "@"); i != -1 {
		urlStr = urlStr[i+1:]
	} else {
		urlStr = strings.TrimPrefix(urlStr, "http://")
		urlStr = strings.TrimPrefix(urlStr, "https://")
		urlStr = strings.TrimPrefix(urlStr, "ssh://")
	}

	// Extract last two path segments (owner/repo).
	parts := strings.FieldsFunc(urlStr, func(r rune) bool {
		return r == '/' || r == ':'
	})
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1], nil
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
