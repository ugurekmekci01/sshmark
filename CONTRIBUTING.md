# Contributing to sshmark

Thanks for your interest in contributing.

## Setup

```bash
make setup
go mod tidy
make test
```

`make setup` installs a **pre-commit** hook that runs `make ci` (fmt, vet, lint, race tests, build).

Install [golangci-lint](https://golangci-lint.run/welcome/install/) (`brew install golangci-lint`).

## Workflow

1. Open an issue or discuss the change.
2. **Write tests first** (TDD). Tests define when a change is complete.
3. Run `make ci` before opening a PR.
4. Keep PRs focused; match existing code style.

## Pull requests

- CI must pass (format, vet, lint, tests).
- No weakening of existing tests.
- User-facing errors must be clear and actionable.

## Code layout

- `cmd/sshmark` — entrypoint
- `internal/cli` — commands (`add`, `open`, …)
- `internal/config` — project bookmarks
- `internal/ssh` — bootstrap + SSH invocation
- `internal/tunnel` — port checks

## Releases

Maintainers tag `v*` to trigger the release workflow (macOS/Linux, amd64/arm64, checksums).
