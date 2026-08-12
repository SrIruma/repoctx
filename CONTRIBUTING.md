# Contributing to repoctx

Thanks for taking the time to contribute! Contributions of all kinds are
welcome: bug reports, feature requests, documentation, and code.

## Repository overview

- `cmd/repoctx/` — CLI entry point.
- `internal/cli/` — cobra commands (`info`, `generate`, `audit`).
- `internal/project/` — repository scanning and the manifest model.
- `internal/adapters/` — one adapter per manifest format (npm, cargo, go,
  pyproject, make, cmake, ruby, composer, maven, gradle).
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
- `repoctx audit . --check` is the CI-gating workflow: it exits non-zero when
  a context file fails, so it can block merges when `AGENTS.md` / `CLAUDE.md`
  drift from the code. This repository gates itself with it.

## Pull request checklist

- [ ] Branch cut from the current `main`.
- [ ] Conventional, atomic commits with a clear message.
- [ ] Tests added or updated for the change.
- [ ] `go vet ./...`, `gofmt -l .`, and `go test ./...` all pass.
- [ ] README and CHANGELOG updated when user-facing behaviour changes.
- [ ] Description explains the what and the why of the change.

## Versioning

repoctx follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Releases use plain tags (`vX.Y.Z`) with the version injected at build time via
`-ldflags` (see `make release VERSION=vX.Y.Z`).

While the project is on `0.x`, any release may break compatibility. Starting
with `1.0.0` the policy is:

| Kind of change | Version bump | Examples |
|---|---|---|
| Breaking | Major (`X.0.0`) | Contract changes — JSON schema or exit-code changes in `info --json` / `audit --json` (see `docs/contract.md`), removed flags or commands, changed defaults that break consumers. |
| Feature | Minor (`x.Y.0`) | New adapters, new flags, new output — anything additive and non-breaking. |
| Fix | Patch (`x.y.Z`) | Bug fixes, doc corrections, internal refactors. |

The machine contract in `docs/contract.md` is frozen from `1.0.0` onward:
changing the JSON schemas or exit codes requires a deliberate major bump.
Golden tests in `internal/cli/testdata` pin the contract output; regenerate
snapshots (`go test ./internal/cli -run TestGoldenJSON -update`) only as part
of that major bump, with a CHANGELOG entry and release note to match.

## Getting help

Open an issue for bugs and feature ideas, or ask questions in the discussions
tab of the repository.
