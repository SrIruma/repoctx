# Contract

This document describes the machine-observable contract of the
`repoctx` CLI: exit codes, JSON schemas and the flags that shape output.
It is the specification that downstream consumers (scripts, CI gates, editors)
can rely on. Every JSON example in this document is verified against real
command output and pinned by golden tests in `internal/cli/testdata`.

The contract is documented but not frozen: repoctx is on 0.x, and breaking
changes are allowed (and will be called out in the changelog) until 1.0.0.

Throughout this document `<repo>` is a placeholder for the absolute path of
the scanned directory; real output emits absolute paths.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. |
| `1` | Any failure. |

Failures that exit `1` include:

- invalid usage or arguments (`--max-depth 0`, unknown flags, too many args);
- an unrecoverable scan error;
- `audit --check` detecting rot in any context file;
- `generate` targeting a file that exists but has no repoctx markers.

Two soft-failure cases deliberately keep the exit code at `0`:

- `info` on a directory with a corrupt manifest: the manifest is kept in the
  report with its `errors` field populated and no facts, and the command
  succeeds;
- `audit` on a directory with no `AGENTS.md` / `CLAUDE.md`: a human message is
  printed on stdout and the command succeeds.

Errors are reported on stderr with an `error:` prefix. Deliberate non-zero
exits that have already printed their output (e.g. `audit --check --json`)
return the code silently.

## Contract flags

| Flag | Commands | Notes |
|---|---|---|
| `--json` | `info`, `audit` | Machine-readable output on stdout. |
| `--check` | `audit` | Exit `1` when any file fails (CI gating). |
| `--file` | `generate` | Context file(s) to update; repeatable. |
| `--dry-run` | `generate` | Preview changes without writing; exit `0`. |
| `--config` | `info`, `generate`, `audit` | Explicit path to `repoctx.toml`. |
| `--max-depth` | `info`, `generate`, `audit` | Scan depth limit (default `6`); must be positive. |
| `--skip-dirs` | `info`, `generate`, `audit` | Extra directories to skip; repeatable. |

`--json` output is indented JSON (two spaces) written to stdout.

## `info --json` schema

```json
{
  "root": "<repo>",
  "manifests": [
    {
      "path": "backend/go.mod",
      "kind": "go",
      "language": "Go",
      "scope": "backend",
      "commands": [
        {"name": "build", "cmd": "go build ./..."}
      ],
      "deps": ["github.com/spf13/cobra"]
    }
  ]
}
```

| Field | Type | Notes |
|---|---|---|
| `root` | string | Absolute path of the scanned directory. |
| `manifests` | array | `null` when the scan finds no manifests. |
| `manifests[].path` | string | Manifest path relative to `root`, POSIX separators. |
| `manifests[].kind` | string | `npm`, `cargo`, `go`, `pyproject`, `make`, `cmake`, `gemfile`, `composer`, `maven`, `gradle`, `meson`. |
| `manifests[].language` | string | Display name; empty when extraction failed. |
| `manifests[].scope` | string | Directory of the manifest (`"."` at the root). |
| `manifests[].commands` | array | `{name, cmd}` pairs; `null` when extraction failed. |
| `manifests[].deps` | array | Omitted when empty. |
| `manifests[].errors` | array | Extraction failures; present only when the manifest could not be parsed. |
| `detected_other` | array | Reserved for detected-but-unsupported manifests; currently never emitted. |

## `audit --json` schema

`audit --json` emits a JSON array, one report per context file:

```json
[
  {
    "file": "<repo>/AGENTS.md",
    "checks": [
      {"name": "commands", "passed": true, "detail": "3 commands claimed"},
      {"name": "paths", "passed": true, "detail": "all referenced paths exist"}
    ],
    "score": 100,
    "passed": true
  }
]
```

| Field | Type | Notes |
|---|---|---|
| `file` | string | Absolute path of the audited context file. |
| `checks` | array | One entry per check (`commands`, `paths`). |
| `checks[].name` | string | Check identifier. |
| `checks[].passed` | boolean | Whether the check found no rot. |
| `checks[].detail` | string | Human summary. |
| `checks[].issues` | array | Rot findings; omitted when the check passes. |
| `checks[].issues[]` | object | `{command?, path?, detail}`; exactly one of `command`/`path` plus `detail`. |
| `score` | integer | `0`–`100`; `100` only when every check passed. |
| `passed` | boolean | `true` exactly when `score == 100`. |

When no `AGENTS.md` / `CLAUDE.md` exists, `audit --json` prints a human
message on stdout (`no context files (AGENTS.md, CLAUDE.md) found in ...`)
and exits `0` — no JSON is emitted.

## Stability policy

- The exact output of `info --json` and `audit --json` is pinned by golden
  tests in `internal/cli/testdata`. Regenerate snapshots with
  `go test ./internal/cli -run TestGoldenJSON -update` only for a deliberate
  contract change, which must come with a CHANGELOG entry and release note.
- Contract changes are allowed on any `0.x` release; starting with `1.x` they
  require a major version bump.
