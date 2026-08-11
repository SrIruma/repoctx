# AGENTS.md

## What is repoctx

repoctx is a Go CLI that keeps AI coding-agent context files (AGENTS.md, CLAUDE.md) truthful. It scans a repository, extracts facts from real code, regenerates the code-derived sections of context files between markers, and audits context files for rot (stale paths, ghost commands).

## Structure

- `cmd/repoctx/` - CLI entry point.
- `internal/cli/` - cobra commands (`info`, `generate`, `audit`) and version injection.
- `internal/project/` - repository scanning and the manifest model.
- `internal/adapters/` - one adapter per manifest format (npm, cargo, go, pyproject, make).
- `internal/markdown/` - marker-aware parsing and rendering of context files.
- `internal/audit/` - rot checks and report scoring.
- `tests/fixtures/` - fixture projects used by tests.

## Conventions

- Conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `chore:`).
- `main` is always releasable; work in feature branches.
- Releases use plain tags (`v0.0.1`), injected into the CLI version variable via `-ldflags`.
- Public-facing code and docs in English; working notes may be in Spanish.
- Commands are atomic and factual: `generate` never touches human-written content outside markers.

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
