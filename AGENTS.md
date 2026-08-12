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
- `internal/adapters/` - one adapter per manifest format (npm, cargo, go, pyproject, make).
- `internal/markdown/` - marker-aware parsing and rendering of context files.
- `internal/audit/` - rot checks and report scoring.
- `tests/fixtures/` - fixture projects used by tests.

## Conventions

- Conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `chore:`).
- `main` is always releasable; work in feature branches.
- Releases use plain tags (`v0.2.0`), injected into `internal/cli.version` via `-ldflags`.
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
| `make install` | `Makefile` |
| `make release` | `Makefile` |
| `make test` | `Makefile` |
| `go build ./...` | `go.mod` |
| `go test ./...` | `go.mod` |
| `go vet ./...` | `go.mod` |
| `gofmt -l .` | `go.mod` |
| `npm run build` | `tests/fixtures/audit/ghost/package.json` |
| `npm run test` | `tests/fixtures/audit/ghost/package.json` |
| `make build` | `tests/fixtures/audit/healthy/Makefile` |
| `make clean` | `tests/fixtures/audit/healthy/Makefile` |
| `make test` | `tests/fixtures/audit/healthy/Makefile` |
| `npm run test` | `tests/fixtures/audit/stale/package.json` |
| `cargo build` | `tests/fixtures/cargo/Cargo.toml` |
| `cargo test` | `tests/fixtures/cargo/Cargo.toml` |
| `cargo run` | `tests/fixtures/cargo/Cargo.toml` |
| `cargo fmt --check` | `tests/fixtures/cargo/Cargo.toml` |
| `cargo clippy` | `tests/fixtures/cargo/Cargo.toml` |
| `go build ./...` | `tests/fixtures/go/go.mod` |
| `go test ./...` | `tests/fixtures/go/go.mod` |
| `go vet ./...` | `tests/fixtures/go/go.mod` |
| `gofmt -l .` | `tests/fixtures/go/go.mod` |
| `make build` | `tests/fixtures/make/Makefile` |
| `make clean` | `tests/fixtures/make/Makefile` |
| `make test` | `tests/fixtures/make/Makefile` |
| `npm run build` | `tests/fixtures/npm/package.json` |
| `npm run lint` | `tests/fixtures/npm/package.json` |
| `npm run test` | `tests/fixtures/npm/package.json` |
| `pytest` | `tests/fixtures/pyproject/pyproject.toml` |
| `ruff check .` | `tests/fixtures/pyproject/pyproject.toml` |
| `go build ./...` | `tests/fixtures/scanner/mono/backend/go.mod` |
| `go test ./...` | `tests/fixtures/scanner/mono/backend/go.mod` |
| `go vet ./...` | `tests/fixtures/scanner/mono/backend/go.mod` |
| `gofmt -l .` | `tests/fixtures/scanner/mono/backend/go.mod` |
| `npm run build` | `tests/fixtures/scanner/mono/package.json` |
| `npm run test` | `tests/fixtures/scanner/mono/package.json` |
| `cargo build` | `tests/fixtures/scanner/mono/tools/rust/Cargo.toml` |
| `cargo test` | `tests/fixtures/scanner/mono/tools/rust/Cargo.toml` |
| `cargo run` | `tests/fixtures/scanner/mono/tools/rust/Cargo.toml` |
| `cargo fmt --check` | `tests/fixtures/scanner/mono/tools/rust/Cargo.toml` |
| `cargo clippy` | `tests/fixtures/scanner/mono/tools/rust/Cargo.toml` |

## Modules

| Module | Language | Dependencies |
|---|---|---|
| `Makefile` | Generic (Make) |  |
| `go.mod` | Go | github.com/inconshreveable/mousetrap, github.com/pelletier/go-toml/v2, github.com/spf13/cobra, github.com/spf13/pflag |
| `tests/fixtures/audit/ghost/package.json` | JavaScript/TypeScript | typescript |
| `tests/fixtures/audit/healthy/Makefile` | Generic (Make) |  |
| `tests/fixtures/audit/stale/package.json` | JavaScript/TypeScript |  |
| `tests/fixtures/cargo/Cargo.toml` | Rust | anyhow, criterion, serde |
| `tests/fixtures/go/go.mod` | Go | github.com/spf13/cobra, github.com/stretchr/testify |
| `tests/fixtures/make/Makefile` | Generic (Make) |  |
| `tests/fixtures/npm/package.json` | JavaScript/TypeScript | react, typescript |
| `tests/fixtures/pyproject/pyproject.toml` | Python | click, requests |
| `tests/fixtures/scanner/mono/backend/go.mod` | Go | github.com/spf13/cobra |
| `tests/fixtures/scanner/mono/package.json` | JavaScript/TypeScript |  |
| `tests/fixtures/scanner/mono/tools/rust/Cargo.toml` | Rust |  |
<!-- repoctx:end -->
