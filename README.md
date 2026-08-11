# repoctx

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

## How it works

1. **Scan** — repoctx walks the repository (skipping `node_modules`, `.git`, build output, …) and detects manifests: `package.json`, `Cargo.toml`, `go.mod`, `pyproject.toml`, `Makefile`. Unsupported ecosystems (`CMakeLists.txt`, `Gemfile`, …) are reported, not ignored.
2. **Extract** — each detected manifest is read by an adapter that turns it into facts: available commands and dependencies.
3. **Generate** — the facts are written between `<!-- repoctx:start -->` / `<!-- repoctx:end -->` markers. Anything outside the markers — your prose, conventions, warnings — is left untouched.
4. **Audit** — checks the claims in your context files against the current state of the code and scores the drift.

## Markers

```markdown
## Commands

<!-- repoctx:start -->
| Command | Description |
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
make release VERSION=v0.1.0   # dist/repoctx_linux_amd64, dist/repoctx_linux_arm64
```

## License

MIT
