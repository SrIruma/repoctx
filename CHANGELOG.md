# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `audit` no longer flags `AGENTS.md` / `CLAUDE.md` or Go symbols such as
  `internal/cli.version` as stale paths. Path detection now uses a known set of
  file extensions while keeping directory detection.

## [v0.1.0] - 2026-08-11

### Added

- `generate` now renders a **Modules** table (`| Module | Language | Dependencies |`)
  next to the Commands table, derived from the detected manifests.
- `ParseCommands` only treats two-cell rows as command rows, so module rows are
  never misread as commands by `audit`.
- Release builds now target `linux/amd64`, `linux/arm64`, and `windows/amd64`
  (Windows binaries ship with a `.exe` suffix).
- Every release ships a `SHA256SUMS.txt` checksum manifest, verified after
  building by `make release`.

### Changed

- This repository's own `AGENTS.md` is now generated and audited by repoctx
  (dogfooding), passing a 100/100 health score and regenerating idempotently.

## [v0.0.1] - 2026-08-11

### Added

- Initial release with the three commands:
  - `repoctx info [dir]` — detect manifests and extract facts.
  - `repoctx generate [dir]` — rewrite code-derived sections between repoctx
    markers, preserving human-written content.
  - `repoctx audit [dir]` — detect ghost commands and stale paths, with a
    health score.
- Markers `<!-- repoctx:start -->` / `<!-- repoctx:end -->` for marker-aware
  parsing and rendering.
- Adapters for `package.json`, `Cargo.toml`, `go.mod`, `pyproject.toml`, and
  `Makefile`. Unsupported ecosystems (CMake, Gemfile, composer, pom, gradle,
  meson) are detected and reported.
- CI workflow (`go vet`, `go build`, `go test`) on push to `main` and PRs.

[Unreleased]: https://github.com/SrIruma/repoctx/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/SrIruma/repoctx/compare/v0.0.1...v0.1.0
[v0.0.1]: https://github.com/SrIruma/repoctx/releases/tag/v0.0.1
