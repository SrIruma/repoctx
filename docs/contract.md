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
- `audit --check` detecting rot in any context file (a *new* finding only,
  when `--since` is given);
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
| `--since <rev>` | `audit` | Incremental audit: only findings on lines changed since `<rev>` fail. |
| `--file` | `generate` | Context file(s) to update; repeatable. |
| `--file` | `audit` | Context file(s) to audit; repeatable. When set, replaces `files` from `repoctx.toml` and the defaults (`AGENTS.md`, `CLAUDE.md`). |
| `--dry-run` | `generate`, `init` | Preview changes without writing; exit `0`. |
| `--force` | `init` | Overwrite an existing context file. |
| `--format` | `skill` | Target agent hint (`claude`, `cursor`, `agents`, `opencode`); output is identical. |
| `--config` | `info`, `generate`, `audit` | Explicit path to `repoctx.toml`. |
| `--max-depth` | `info`, `generate`, `audit` | Scan depth limit (default `6`); must be positive. |
| `--skip-dirs` | `info`, `generate`, `audit` | Extra directories to skip; repeatable. |

`--json` output is indented JSON (two spaces) written to stdout.

## Fleet mode (multi-root audit)

`audit` accepts one or more roots: `repoctx audit <dir>...`. With a single root
the output shape is unchanged (array of per-file reports). With two or more
roots the JSON shape becomes:

```json
{
  "roots": ["<repo-a>", "<repo-b>"],
  "reports": [
    { "root": "<repo-a>", "files": [ { "file": ..., "checks": ..., "score": ..., "passed": ... } ] }
  ],
  "fleet_score": 100,
  "passed": true
}
```

A single root whose `repoctx.toml` contains a `roots = [...]` list expands to
fleet mode automatically (paths resolved relative to the config file). Exit code
is `1` when any file in any root fails (`--check`).

## Incremental audit (`--since`)

`audit --since <rev>` compares the audited context files against a git revision
and classifies every finding by whether the line it came from was added or
changed since `<rev>`:

- **new** — the finding's line is part of the diff since `<rev>`. Only new
  findings fail the run.
- **pre-existing** — the finding exists but predates `<rev>`. It is reported
  but does not fail the run.

This lets a fleet adopt `--check` in CI without first fixing every repository:
the gate fires only on drift the change introduced. `<rev>` is anything
`git rev-parse` accepts (`HEAD`, a branch, a sha, `origin/main`). A file absent
from `<rev>` (added, renamed, or untracked) counts as entirely new.

Without `--since`, behaviour is unchanged: every finding fails. `--since`
requires the audited root to be inside a git working tree; it errors otherwise,
and errors when `<rev>` cannot be resolved. The engine itself never shells out
to git — the CLI computes the changed-line set (`internal/cli/changes.go`) and
passes it to the pure `internal/audit` package.

`init` writes the skeleton to the target file (or stdout with `--dry-run`). `skill` writes the canonical `SKILL.md` to stdout; its output is stable and deterministic.

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
| `manifests[].canonical` | boolean | `false` when the manifest lives under a test/fixture/example/vendor directory (see [Canonical scoping](#canonical-scoping)); such manifests are reported but excluded from the generated tables. |
| `manifests[].commands` | array | `{name, cmd}` pairs; `null` when extraction failed. |
| `manifests[].package_manager` | string | npm-kind only: `npm`, `yarn`, `pnpm` or `bun` — the tool that runs the package.json scripts. Detected from the corepack `packageManager` field, then from a sibling lockfile (`yarn.lock`, `pnpm-lock.yaml`, `bun.lock`, `bun.lockb`, `package-lock.json`, `npm-shrinkwrap.json`); `npm` when there is no signal. Commands use the matching prefix (`yarn run build`, `pnpm run test`, `bun run lint`). |
| `manifests[].deps` | array | Omitted when empty. |
| `manifests[].errors` | array | Extraction failures; present only when the manifest could not be parsed. |
| `context_files` | array | Context files discovered in the tree, relative to `root`, sorted; omitted when none. See [Which files are audited](#which-files-are-audited). |
| `detected_other` | array | Reserved for detected-but-unsupported manifests; currently never emitted. |

## Canonical scoping

A repository often contains manifests that are not part of its real surface:
test fixtures, examples, vendored dependencies. repoctx detects them but marks
them `"canonical": false` when any directory segment of the path is a
recognised non-canonical name (`test`, `tests`, `testdata`, `fixture`,
`fixtures`, `example`, `examples`, `vendor`, `third_party`, `node_modules`,
`__tests__`, `e2e`, `spec`, `specs`).

Non-canonical manifests stay in `info --json` (so nothing is hidden), but they
are excluded from the generated command and module tables, and context files
below non-canonical directories are not discovered. This keeps a repository
like repoctx — which ships dozens of fixture manifests — from drowning its own
context file in fixture noise.

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
| `checks[].passed` | boolean | Whether the check found no rot. With `--since`, a check passes when it has no *new* findings, even if pre-existing ones remain. |
| `checks[].detail` | string | Human summary. |
| `checks[].issues` | array | Rot findings; omitted when the check passes. |
| `checks[].issues[]` | object | `{command?, path?, detail, line?, scope?}`; exactly one of `command`/`path` plus `detail`. `line` is the 1-based context-file line the finding came from. `scope` is present only with `--since` and is `new` or `pre-existing`. |
| `score` | integer | `0`–`100`; `100` only when every check passed. |
| `passed` | boolean | `true` exactly when `score == 100`. |
| `new_issues` | integer | Only with `--since`: findings introduced since the base. Omitted when zero. |
| `pre_existing` | integer | Only with `--since`: inherited findings, reported but not gating. Omitted when zero. |

In fleet mode the top-level object additionally carries `since`, `new_issues`
and `pre_existing` (echoing the base and the totals) when `--since` is used.

### Which files are audited

`audit` and `generate` choose their context files with the same precedence:

1. an explicit `--file` flag (or the `files` list in `repoctx.toml`) wins
   outright: exactly those names are used and the tree is not searched;
2. otherwise the context files discovered in the tree are used, so a monorepo
   keeps one context file per package (a nested `AGENTS.md` owns its own
   subtree, with tables written relative to it);
3. when discovery finds nothing, `audit` falls back to `AGENTS.md` and
   `CLAUDE.md`, and `generate` falls back to creating `AGENTS.md`.

A project can name its context files explicitly:

```toml
# repoctx.toml
files = ["CONTEXT.md"]
```

With `files` set, discovery is disabled and exactly those files are used, so
the set of audited files only differs for projects that configure it.

Path references inside a context file are resolved relative to the file's own
directory first, then relative to the repository root. A nested `AGENTS.md` can
therefore cite `go.mod` (its sibling) or `internal/cli/audit.go` (root-relative)
and both resolve correctly.

When no audited context file exists, `audit --json` prints a human message on
stdout (`no context files (AGENTS.md, CLAUDE.md) found in ...`) and exits `0` —
no JSON is emitted. The message names the files that were looked for, so a
configured project sees `no context files (CONTEXT.md) found in ...`.

## Stability policy

- The exact output of `info --json` and `audit --json` is pinned by golden
  tests in `internal/cli/testdata`. Regenerate snapshots with
  `go test ./internal/cli -run TestGoldenJSON -update` only for a deliberate
  contract change, which must come with a CHANGELOG entry and release note.
- Contract changes are allowed on any `0.x` release; starting with `1.x` they
  require a major version bump.
