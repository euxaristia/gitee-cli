# gitee-cli

A full-featured Gitee CLI inspired by `gh` for GitHub.

## Features

- Authentication: `auth login/logout/status/token`
- Repository workflows: `repo list/view/create/clone`
- Issue workflows: `issue list/view/create/comment/close/reopen`
- Pull request workflows: `pr list/view/create/view/comment/merge/close`
- Release workflows: `release list/view/create/delete`
- Raw API access: `api`
- Config management: `config list/get/set/unset/path`
- Shell completion: `completion bash|zsh|fish|powershell`
- Output formats: table or JSON (`-o json`)

## Install

```bash
make install
```

This installs to `~/.local/bin/gitee` by default.

## Quick Start

```bash
gitee auth login --token <PAT>
gitee repo list
gitee issue list --repo owner/repo
gitee pr create --repo owner/repo --title "feat: ..." --head feature --base master --body "..."
```

## Authentication

The CLI uses either:

1. `GITEE_TOKEN` environment variable (highest priority)
2. Stored token in OS keychain (`gitee-cli` service)
3. Legacy fallback: `~/.config/gitee-cli/config.yaml` token field (read-only compatibility path)

Generate a personal access token from Gitee account settings and run:

```bash
gitee auth login
```

Security behavior:

- Interactive token input is hidden (no terminal echo).
- API requests use `Authorization` headers (not token-in-query).
- `gitee auth login/logout` manages secrets in keychain.

## Examples

```bash
# Repositories
gitee repo list --org my-org --visibility public
gitee repo view owner/repo
gitee repo create --name new-repo --description "demo" --private

# Issues
gitee issue create --repo owner/repo --title "Bug report" --body "details"
gitee issue comment --repo owner/repo 123 --body "working on this"

# Pull requests
gitee pr list --repo owner/repo --state open
gitee pr merge --repo owner/repo 45 --message "merge PR #45"

# Releases
gitee release create --repo owner/repo --tag v1.2.3 --name "v1.2.3" --body "changelog"

# Raw API
gitee api repos/owner/repo -X GET
gitee api repos/owner/repo/issues -X POST -F title=hello -F body=world
```

## Config

```bash
gitee config list
gitee config set output json
gitee config get api_base
```

## Notes

- Some Gitee API endpoints require account/org-specific permissions.
- `api -H/--header` is accepted for CLI parity and reserved for future custom-header passthrough.
