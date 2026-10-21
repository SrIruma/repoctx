# AGENTS.md

## What is repoctx

repoctx keeps the context files your AI agents read honest. Instead of trusting
a hand-maintained `AGENTS.md` or `CLAUDE.md` that slowly drifts out of date, it
scans the real code, extracts the commands and module structure as facts, and
regenerates those code-derived sections between its markers. It also audits the
file for rot — ghost commands and stale paths — so you always know when the
context is lying.

## Structure

The code is split into small, single-purpose packages:

- `cmd/repoctx/` - CLI entry point.
- `internal/cli/` - cobra commands (`info`, `generate`, `audit`, `workflow`) and version injection.
- `internal/project/` - repository scanning and the manifest model.
- `internal/adapters/` - one adapter per manifest format (npm, cargo, go, pyproject, make, cmake, ruby, composer, maven, gradle).
- `internal/markdown/` - marker-aware parsing and rendering of context files.
- `internal/audit/` - rot checks and report scoring.
- `tests/fixtures/` - fixture projects used by tests.

## Conventions

- Conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `chore:`).
- `main` is always releasable; work in feature branches.
- Releases use plain tags (`vX.Y.Z`), injected into `internal/cli.version` via `-ldflags`.
- Public-facing code and docs in English; working notes may be in Spanish.
- Commands are atomic and factual: `generate` never touches human-written content outside markers.
- This repository is dogfooded: the Commands/Modules tables below are generated
  by repoctx between the markers — never hand-edit them. After changing
  scripts, manifests, dependencies, or code layout, regenerate with
  `repoctx generate .` and verify with `repoctx audit . --check`.
- Development quickstart: `make test`, `make build`; before committing run
  `go vet ./...` and `gofmt -l .`. See CONTRIBUTING.md.
- In the Commands table, "Source" is the manifest that triggered the adapter —
  e.g. `go vet ./...` is an adapter convention, not a literal `go.mod` key.

<!-- repoctx:start -->
## Commands

| Command | Source |
|---|---|
| `make all` | `Makefile` |
| `make build` | `Makefile` |
| `make clean` | `Makefile` |
| `make docs-examples` | `Makefile` |
| `make install` | `Makefile` |
| `make release` | `Makefile` |
| `make test` | `Makefile` |
| `go build ./...` | `go.mod` |
| `go test ./...` | `go.mod` |
| `go vet ./...` | `go.mod` |
| `gofmt -l .` | `go.mod` |

## Modules

| Module | Language | Dependencies |
|---|---|---|
| `Makefile` | Generic (Make) |  |
| `go.mod` | Go | github.com/inconshreveable/mousetrap, github.com/pelletier/go-toml/v2, github.com/spf13/cobra, github.com/spf13/pflag |
<!-- repoctx:end -->
