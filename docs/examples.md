# Usage examples

This cookbook shows repoctx in action across every relevant case: every
command, every output format, the failure modes, the scan-tuning flags and the
workspace layouts. It is **generated** — see `scripts/gen-examples.sh` — and
every block below is the real output of the CLI run against the fixture
projects in `tests/fixtures/`, with absolute paths normalized to the `<repo>`
placeholder. If it ever drifts from reality, `make docs-examples` (verified in
CI) will fail.

See [contract.md](contract.md) for the machine contract (exit codes, JSON
schemas, contract flags).

## Commands at a glance

| Command | Purpose |
|---|---|
| `repoctx info [dir]` | Detect manifests and extract facts. |
| `repoctx generate [dir]` | Rewrite the code-derived sections of a context file between repoctx markers. |
| `repoctx audit [dir]` | Detect context rot with a health score. |
| `repoctx workflow [file ...]` | Print a paste-ready block for keeping a context file truthful. |

## info

Detect manifests and extract facts from a repository. `info` prints a human
table by default; add `--json` for the machine-readable contract.

Monorepo with a Go backend, a `package.json` and a Rust tool:

```console
$ repoctx info tests/fixtures/scanner/mono
Detected manifests in <repo>/scanner/mono:
  backend/go.mod               go         Go                     commands: [build, test, vet, fmt]  (1 deps)
  package.json                 npm        JavaScript/TypeScript  commands: [build, test]  (0 deps)
  tools/rust/Cargo.toml        cargo      Rust                   commands: [build, test, run, fmt, clippy]  (0 deps)
```

Same project, JSON output (the `root` field and `manifest[].path` are part of
the documented contract):

```json
$ repoctx info tests/fixtures/scanner/mono --json
{
  "root": "<repo>/scanner/mono",
  "manifests": [
    {
      "path": "backend/go.mod",
      "kind": "go",
      "language": "Go",
      "scope": "backend",
      "commands": [
        {
          "name": "build",
          "cmd": "go build ./..."
        },
        {
          "name": "test",
          "cmd": "go test ./..."
        },
        {
          "name": "vet",
          "cmd": "go vet ./..."
        },
        {
          "name": "fmt",
          "cmd": "gofmt -l ."
        }
      ],
      "deps": [
        "github.com/spf13/cobra"
      ]
    },
    {
      "path": "package.json",
      "kind": "npm",
      "language": "JavaScript/TypeScript",
      "scope": ".",
      "commands": [
        {
          "name": "build",
          "cmd": "npm run build"
        },
        {
          "name": "test",
          "cmd": "npm run test"
        }
      ],
      "package_manager": "npm"
    },
    {
      "path": "tools/rust/Cargo.toml",
      "kind": "cargo",
      "language": "Rust",
      "scope": "tools/rust",
      "commands": [
        {
          "name": "build",
          "cmd": "cargo build"
        },
        {
          "name": "test",
          "cmd": "cargo test"
        },
        {
          "name": "run",
          "cmd": "cargo run"
        },
        {
          "name": "fmt",
          "cmd": "cargo fmt --check"
        },
        {
          "name": "clippy",
          "cmd": "cargo clippy"
        }
      ]
    }
  ]
}
```

A corrupt manifest is a soft failure: the scan keeps the broken manifest with
zero facts and reports why (`!` in human output, `errors` in JSON), and the
exit code stays `0`. Here `broken/package.json` is not valid JSON:

```console
$ repoctx info tests/fixtures/scanner/corrupt
Detected manifests in <repo>/scanner/corrupt:
  broken/package.json          npm                               commands: []  (0 deps)
  ! broken/package.json: invalid character 'n' looking for beginning of object key string
  package.json                 npm        JavaScript/TypeScript  commands: [test]  (1 deps)
```

```json
$ repoctx info tests/fixtures/scanner/corrupt --json
{
  "root": "<repo>/scanner/corrupt",
  "manifests": [
    {
      "path": "broken/package.json",
      "kind": "npm",
      "language": "",
      "scope": "broken",
      "commands": null,
      "errors": [
        "invalid character 'n' looking for beginning of object key string"
      ]
    },
    {
      "path": "package.json",
      "kind": "npm",
      "language": "JavaScript/TypeScript",
      "scope": ".",
      "commands": [
        {
          "name": "test",
          "cmd": "npm run test"
        }
      ],
      "deps": [
        "typescript"
      ],
      "package_manager": "npm"
    }
  ]
}
```

## generate

`generate` rewrites only the content between the `<!-- repoctx:start -->` and
`<!-- repoctx:end -->` markers, preserving human-written prose. `--dry-run`
previews what would change without writing anything, and exits `0` either way.

Here the `ghost` fixture claims a `lint` script that was removed from its
`package.json`, so a generation would rewrite its Commands table:

```console
$ repoctx generate tests/fixtures/audit/ghost --dry-run
would update AGENTS.md
```

Once a context file is fully in sync, the same dry-run reports it as current:

```console
$ repoctx generate <fresh copy of healthy> --dry-run
AGENTS.md is up to date
```

## audit

`audit` checks a context file's claims against the current code and scores the
drift. A fully in-sync file passes 100/100:

```console
$ repoctx audit tests/fixtures/audit/healthy
PASS  <repo>/audit/healthy/AGENTS.md  score 100/100
  ok  commands: 3 commands claimed
  ok  paths: all referenced paths exist
```

Rot is reported per check — a stale path here, a ghost command there:

```console
$ repoctx audit tests/fixtures/audit/stale
FAIL  <repo>/audit/stale/AGENTS.md  score 50/100
  ok  commands: 1 commands claimed
  !!  paths: 1 stale path(s)
    - legacy/schema.md: stale path: not found in repository
```

```console
$ repoctx audit tests/fixtures/audit/ghost
FAIL  <repo>/audit/ghost/AGENTS.md  score 50/100
  !!  commands: 1 ghost command(s) out of 2 claimed
    - npm run lint: ghost command: no longer present in the project
  ok  paths: all referenced paths exist
```

JSON output, one report per context file:

```json
$ repoctx audit tests/fixtures/audit/healthy --json
[
  {
    "file": "<repo>/audit/healthy/AGENTS.md",
    "checks": [
      {
        "name": "commands",
        "passed": true,
        "detail": "3 commands claimed"
      },
      {
        "name": "paths",
        "passed": true,
        "detail": "all referenced paths exist"
      }
    ],
    "score": 100,
    "passed": true
  }
]
```

## Exit codes

`0` means success, `1` means any failure. `audit --check` turns rot into an
exit code so CI can gate merges — healthy passes, rotting context fails:

```console
$ repoctx audit tests/fixtures/audit/healthy --check; echo $?
PASS  <repo>/audit/healthy/AGENTS.md  score 100/100
  ok  commands: 3 commands claimed
  ok  paths: all referenced paths exist
0
```

```console
$ repoctx audit tests/fixtures/audit/ghost --check; echo $?
FAIL  <repo>/audit/ghost/AGENTS.md  score 50/100
  !!  commands: 1 ghost command(s) out of 2 claimed
    - npm run lint: ghost command: no longer present in the project
  ok  paths: all referenced paths exist
1
```

Invalid usage is a failure too:

```console
$ repoctx info tests/fixtures/scanner/mono --max-depth 0; echo $?
error: --max-depth must be a positive integer
1
```

## Scan tuning flags

`--max-depth` limits how far the scanner descends; `--skip-dirs` (repeatable)
adds directories to skip on top of the built-ins (`.git`, `node_modules`,
`vendor`, ...). The same flags work on `info`, `generate` and `audit`.

`--max-depth 1` on the monorepo keeps only the root manifest:

```console
$ repoctx info tests/fixtures/scanner/mono --max-depth 1
Detected manifests in <repo>/scanner/mono:
  package.json                 npm        JavaScript/TypeScript  commands: [build, test]  (0 deps)
```

`--skip-dirs tools` drops the Rust tool directory:

```console
$ repoctx info tests/fixtures/scanner/mono --skip-dirs tools
Detected manifests in <repo>/scanner/mono:
  backend/go.mod               go         Go                     commands: [build, test, vet, fmt]  (1 deps)
  package.json                 npm        JavaScript/TypeScript  commands: [build, test]  (0 deps)
```

The same tuning lives declaratively in a `repoctx.toml` in the scanned
directory, or in any location via `--config`. Precedence is
`flags > config > defaults`:

```console
$ repoctx info tests/fixtures/scanner/mono --config <path to repoctx.toml>
Detected manifests in <repo>/scanner/mono:
  backend/go.mod               go         Go                     commands: [build, test, vet, fmt]  (1 deps)
  package.json                 npm        JavaScript/TypeScript  commands: [build, test]  (0 deps)
```

## workflow

`workflow` prints a paste-ready block for the human-written section of a
context file, telling agents to regenerate and gate on truth:

```console
$ repoctx workflow
## Repo Context Maintenance

The Commands and Modules tables between `<!-- repoctx:start -->` and
`<!-- repoctx:end -->` in `AGENTS.md` are generated by repoctx from the code —
never hand-edit them.

After you change commands, scripts, dependencies, manifests, or move packages, run:
1. `repoctx generate .`  — regenerate the code-derived tables
2. `repoctx audit . --check`  — fail if the context drifted

Fix anything `repoctx audit` reports before finishing the task.
```

## Workspaces

Workspace (monorepo) layouts are scanned naturally: every manifest in the tree
is detected, commands are attributed per manifest, and identical commands from
different packages are disambiguated by the Source column.

npm/pnpm workspace with a root `package.json` and two member packages:

```json
$ repoctx info tests/fixtures/scanner/workspace-npm --json
{
  "root": "<repo>/scanner/workspace-npm",
  "manifests": [
    {
      "path": "package.json",
      "kind": "npm",
      "language": "JavaScript/TypeScript",
      "scope": ".",
      "commands": [
        {
          "name": "build",
          "cmd": "npm run build"
        },
        {
          "name": "test",
          "cmd": "npm run test"
        }
      ],
      "package_manager": "npm"
    },
    {
      "path": "packages/app/package.json",
      "kind": "npm",
      "language": "JavaScript/TypeScript",
      "scope": "packages/app",
      "commands": [
        {
          "name": "test",
          "cmd": "npm run test"
        }
      ],
      "deps": [
        "react"
      ],
      "package_manager": "npm"
    },
    {
      "path": "packages/lib/package.json",
      "kind": "npm",
      "language": "JavaScript/TypeScript",
      "scope": "packages/lib",
      "commands": [
        {
          "name": "test",
          "cmd": "npm run test"
        }
      ],
      "deps": [
        "typescript"
      ],
      "package_manager": "npm"
    }
  ]
}
```

Cargo workspace: a virtual root (no `[package]`) exposes the workspace-wide
commands but not `cargo run`, which has no default binary; member crates keep
the full set. `[workspace.dependencies]` are extracted too:

```json
$ repoctx info tests/fixtures/scanner/workspace-cargo --json
{
  "root": "<repo>/scanner/workspace-cargo",
  "manifests": [
    {
      "path": "Cargo.toml",
      "kind": "cargo",
      "language": "Rust",
      "scope": ".",
      "commands": [
        {
          "name": "build",
          "cmd": "cargo build"
        },
        {
          "name": "test",
          "cmd": "cargo test"
        },
        {
          "name": "fmt",
          "cmd": "cargo fmt --check"
        },
        {
          "name": "clippy",
          "cmd": "cargo clippy"
        }
      ],
      "deps": [
        "anyhow",
        "serde"
      ]
    },
    {
      "path": "crates/a/Cargo.toml",
      "kind": "cargo",
      "language": "Rust",
      "scope": "crates/a",
      "commands": [
        {
          "name": "build",
          "cmd": "cargo build"
        },
        {
          "name": "test",
          "cmd": "cargo test"
        },
        {
          "name": "run",
          "cmd": "cargo run"
        },
        {
          "name": "fmt",
          "cmd": "cargo fmt --check"
        },
        {
          "name": "clippy",
          "cmd": "cargo clippy"
        }
      ],
      "deps": [
        "serde",
        "tokio"
      ]
    },
    {
      "path": "crates/b/Cargo.toml",
      "kind": "cargo",
      "language": "Rust",
      "scope": "crates/b",
      "commands": [
        {
          "name": "build",
          "cmd": "cargo build"
        },
        {
          "name": "test",
          "cmd": "cargo test"
        },
        {
          "name": "run",
          "cmd": "cargo run"
        },
        {
          "name": "fmt",
          "cmd": "cargo fmt --check"
        },
        {
          "name": "clippy",
          "cmd": "cargo clippy"
        }
      ],
      "deps": [
        "anyhow",
        "thiserror"
      ]
    }
  ]
}
```

## Use in CI

Gate merges on context truth. `--check` exits non-zero when any audited file
fails, so a job can block a pull request when context files drift from the
code:

```yaml
# .github/workflows/ci.yml
- run: repoctx audit . --check
```

It composes with `--json` for machine-readable failures:

```yaml
- run: repoctx audit . --check --json
```
