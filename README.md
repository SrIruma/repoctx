# repoctx

[![Go Version](https://img.shields.io/github/go-mod/go-version/SrIruma/repoctx)](https://go.dev/doc/install)
[![CI](https://img.shields.io/github/actions/workflow/status/SrIruma/repoctx/ci.yml)](https://github.com/SrIruma/repoctx/actions)
[![Release](https://img.shields.io/github/v/release/SrIruma/repoctx)](https://github.com/SrIruma/repoctx/releases)
[![Downloads](https://img.shields.io/github/downloads/SrIruma/repoctx/total)](https://github.com/SrIruma/repoctx/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/SrIruma/repoctx)](https://goreportcard.com/report/github.com/SrIruma/repoctx)
[![License](https://img.shields.io/github/license/SrIruma/repoctx)](LICENSE)

repoctx keeps your AI coding-agent context files (`AGENTS.md`, `CLAUDE.md`) truthful. It scans a repository, extracts facts from the real code — build/test commands, module structure, dependencies — and uses them to regenerate the factual sections of your context files, preserving anything a human wrote.

## Why

Context files rot. A README mentions a `make release` target that no longer exists; `AGENTS.md` lists `backend/` as a path that was renamed months ago. Agents read those files and act on stale information. repoctx is the source of truth for everything that can be derived from the repository itself.

## Install

```sh
go install github.com/SrIruma/repoctx/cmd/repoctx@latest
```

Or build from source:

```sh
make build   # -> bin/repoctx
make install # -> $GOBIN/repoctx
```

## Commands

| Command | Description |
|---|---|
| `repoctx info [dir]` | Detect manifests and extract facts from a repository. |
| `repoctx generate [dir]` | Regenerate the code-derived sections of `AGENTS.md` / `CLAUDE.md` between repoctx markers. |
| `repoctx audit [dir]` | Detect context rot: stale paths and ghost commands, with a health score. |

Run `repoctx <command> --help` for full usage. Every command supports `--json`.

## Demo

Inspect what a monorepo looks like to repoctx:

```console
$ repoctx info tests/fixtures/scanner/mono
Detected manifests in /path/to/scanner/mono:
  backend/go.mod               go         Go                     commands: [build, test, vet, fmt]  (1 deps)
  package.json                 npm        JavaScript/TypeScript  commands: [build, test]  (0 deps)
  tools/rust/Cargo.toml        cargo      Rust                   commands: [build, test, run, fmt, clippy]  (0 deps)
```

Audit a context file and get a health score:

```console
$ repoctx audit tests/fixtures/audit/healthy
PASS  /path/to/audit/healthy/AGENTS.md  score 100/100
  ok  commands: 3 commands claimed
  ok  paths: all referenced paths exist
```

## How it works

1. **Scan** — repoctx walks the repository (skipping `node_modules`, `.git`, build output, …) and detects manifests: `package.json`, `Cargo.toml`, `go.mod`, `pyproject.toml`, `Makefile`. Unsupported ecosystems (`CMakeLists.txt`, `Gemfile`, …) are reported, not ignored.
2. **Extract** — each detected manifest is read by an adapter that turns it into facts: available commands and dependencies.
3. **Generate** — the facts are written between `<!-- repoctx:start -->` / `<!-- repoctx:end -->` markers. Anything outside the markers — your prose, conventions, warnings — is left untouched.
4. **Audit** — checks the claims in your context files against the current state of the code and scores the drift.

## Markers

```markdown
<!-- repoctx:start -->
| Command | Source |
|---|---|
| `make test` | Run the test suite |
<!-- repoctx:end -->
```

`repoctx generate` rewrites only the content between the markers. Files without markers are never written.

## Development

```sh
make test      # go test ./...
make build     # bin/repoctx
make install   # $GOBIN/repoctx
make release VERSION=v0.1.0   # dist/repoctx_linux_amd64, repoctx_linux_arm64, repoctx_windows_amd64.exe + SHA256SUMS.txt
```

## Community

- [Contributing](CONTRIBUTING.md) — how to set up, branch, commit, and test.
- [Changelog](CHANGELOG.md) — release history.
- [Code of Conduct](CODE_OF_CONDUCT.md) — community guidelines.
- [Issues](https://github.com/SrIruma/repoctx/issues) — report bugs and request features.

## License

[MIT](LICENSE)
