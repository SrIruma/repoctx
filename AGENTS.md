# AGENTS.md

## What is repoctx

repoctx is a Go CLI that keeps AI coding-agent context files (`AGENTS.md`, `CLAUDE.md`) truthful. It scans a repository, extracts facts from real code, regenerates the code-derived sections of context files between markers, and audits context files for rot (stale paths, ghost commands).

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
- Releases use plain tags (`v0.0.1`), injected into `internal/cli.version` via `-ldflags`.
- Public-facing code and docs in English; working notes may be in Spanish.
- Commands are atomic and factual: `generate` never touches human-written content outside markers.
