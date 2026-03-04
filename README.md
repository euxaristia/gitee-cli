# gitee-cli

A full-featured Gitee CLI inspired by `gh` for GitHub.

## Features

- Authentication: `auth login/logout/status/token`
- Repository workflows: `repo list/view/create/clone`
- Git wrappers with retry for unstable network: `git push/pull`, and shortcuts `push/pull/commit`
- Issue workflows: `issue list/view/create/comment/close/reopen`
- Pull request workflows: `pr list/view/create/view/comment/merge/close`
- Release workflows: `release list/view/create/delete`
- Raw API access: `api`
- Config management: `config list/get/set/unset/path`
- Shell completion: `completion bash|zsh|fish|powershell`
- Output formats: table or JSON (`-o json`)

## Install

```bash
go install ./cmd/gitee
```

This installs the `gitee` binary to your `$GOPATH/bin` (usually `~/go/bin`).

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

## Implementation Details

### Architecture
- **CLI Framework**: Built using [Cobra](https://github.com/spf13/cobra), providing a structured and discoverable command hierarchy.
- **Unified State**: A central `App` struct is initialized at startup, ensuring consistent configuration and API client settings (base URL, timeouts, etc.) across all subcommands.

### Resilience & Retries
Designed specifically for high reliability over unstable networks:
- **API Retries**: The internal HTTP client automatically retries idempotent requests (GET, HEAD) up to 3 times with exponential backoff if transient network errors (e.g., connection resets, TLS timeouts) are detected.
- **Git Wrappers**: Commands like `gitee push` and `gitee pull` wrap the native `git` binary, monitoring stderr for common network failures and automatically retrying the operation.

### Security
- **Credential Storage**: Uses [go-keyring](https://github.com/zalando/go-keyring) to store access tokens in the system's native secure store (macOS Keychain, Linux Secret Service, or Windows Credential Manager) rather than plaintext files.
- **Priority**: Tokens are resolved in order: `GITEE_TOKEN` environment variable > OS Keychain > Legacy config file.

## Examples

```bash
# Repositories
gitee repo list --org my-org --visibility public
gitee repo view owner/repo
gitee repo create --name new-repo --description "demo" --private

# Git wrappers
gitee push origin master
gitee pull --rebase
gitee commit -m "feat: improve retry behavior"

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
