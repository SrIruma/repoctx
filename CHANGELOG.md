# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- New `repoctx.toml` config file: `max_depth`, `skip_dirs` and `files` tune
  scanning and the default context files without per-invocation flags. Loaded
  from the target directory, or from an explicit path via `--config`. Values
  merge with the CLI flags in `flags > config > defaults` precedence order.
- `--max-depth` flag on `info`, `generate` and `audit` to limit how far the
  scanner descends (default 6; a non-positive value is an error). The depth
  limit now also applies to files, not only to directory descent.
- `--skip-dirs` (repeatable) flag on `info`, `generate` and `audit` to add
  directories to the skip list; built-in skips (`.git`, `node_modules`,
  `vendor`, ...) are always preserved.
- `generate` uses the `files` list from `repoctx.toml` when `--file` is not
  given (falling back to `AGENTS.md`).

## [v0.3.0] - 2026-08-12

### Added

- New adapters for five more ecosystems: CMake (`CMakeLists.txt`, custom
  targets as `cmake --build build --target <name>` commands, `find_package`
  dependencies), Ruby (`Gemfile` gems as `bundle exec <gem>` commands and
  dependencies), PHP (`composer.json` scripts as `composer run <script>`
  commands plus `require`/`require-dev` dependencies), Java (`pom.xml`
  `groupId:artifactId[:version]` dependencies) and Gradle (`build.gradle` /
  `build.gradle.kts` dependencies from `implementation`, `api`, `compileOnly`,
  `runtimeOnly`, `testImplementation`, `annotationProcessor`, `kapt` and
  `classpath` configurations). The scanner now reports only `meson.build` as
  detected-but-unsupported.

## [v0.2.0] - 2026-08-12

### Added

- `repoctx audit --check` exits non-zero when any audited file fails, turning
  the health report into a CI gate that blocks merges when context files rot.
  Works combined with `--json` for machine-readable failures.
- `repoctx workflow [file ...]` prints a paste-ready block for the
  human-written section of a context file. It tells coding agents to regenerate
  the repoctx tables after code changes and to gate on
  `repoctx audit . --check`, so the file stays truthful by convention, not by
  luck.

### Fixed

- `audit` no longer flags `AGENTS.md` / `CLAUDE.md` or Go symbols such as
  `internal/cli.version` as stale paths. Path detection now uses a known set of
  file extensions while keeping directory detection.

### Changed

- The repo's own CI now runs `repoctx audit . --check`, so `AGENTS.md` cannot
  rot on `main` (dogfooding the CI gate).

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

[Unreleased]: https://github.com/SrIruma/repoctx/compare/v0.3.0...HEAD
[v0.3.0]: https://github.com/SrIruma/repoctx/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/SrIruma/repoctx/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/SrIruma/repoctx/compare/v0.0.1...v0.1.0
[v0.0.1]: https://github.com/SrIruma/repoctx/releases/tag/v0.0.1
