# Contributing to repoctx

Thanks for taking the time to contribute! Contributions of all kinds are
welcome: bug reports, feature requests, documentation, and code.

## Repository overview

- `cmd/repoctx/` — CLI entry point.
- `internal/cli/` — cobra commands (`info`, `generate`, `audit`).
- `internal/project/` — repository scanning and the manifest model.
- `internal/adapters/` — one adapter per manifest format (npm, cargo, go,
  pyproject, make).
- `internal/markdown/` — marker-aware parsing and rendering of context files.
- `internal/audit/` — rot checks and report scoring.
- `tests/fixtures/` — fixture projects used by the test suite.

## Development environment

- Go 1.22 or newer.
- A POSIX `make`.

```sh
make build     # bin/repoctx (local dev binary)
make test      # go test ./...
make install   # go install -> $GOBIN/repoctx
make release VERSION=v0.1.0   # dist/ binaries + SHA256SUMS.txt
```

Before committing, make sure the basics are green:

```sh
go vet ./...
gofmt -l .
go test ./...
```

## Branching and commits

- Work on a feature branch named after the change (for example
  `feat/<something>`, `fix/<something>`, `docs/<something>`).
- Keep `main` always releasable: merge to `main` only through a pull request.
- Use [conventional commits](https://www.conventionalcommits.org/) and keep
  each commit atomic:

| Prefix  | When to use                          |
|---------|--------------------------------------|
| `feat:` | A new user-facing capability.        |
| `fix:`  | A bug fix.                           |
| `docs:` | Documentation-only changes.          |
| `test:` | Adding or updating tests.            |
| `chore:`| Tooling, build, or maintenance work. |

## Testing conventions

- Adapters and scanning behaviour are covered by fixture projects under
  `tests/fixtures/` (one directory per ecosystem, plus `scanner/mono` for
  monorepos and `audit/{ghost,stale,healthy}` for the rot checks).
- When adding an adapter or changing extraction logic, add or update fixtures
  and tests in the same commit.
- `audit` output is written through `cmd.OutOrStdout()` so CLI tests can
  capture it; `info --json` uses `os.Stdout`.

## Pull request checklist

- [ ] Branch cut from the current `main`.
- [ ] Conventional, atomic commits with a clear message.
- [ ] Tests added or updated for the change.
- [ ] `go vet ./...`, `gofmt -l .`, and `go test ./...` all pass.
- [ ] README and CHANGELOG updated when user-facing behaviour changes.
- [ ] Description explains the what and the why of the change.

## Getting help

Open an issue for bugs and feature ideas, or ask questions in the discussions
tab of the repository.
