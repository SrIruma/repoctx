#!/usr/bin/env bash
#
# Regenerates docs/examples.md from the real CLI output.
#
# Every example in the cookbook is produced by running bin/repoctx against the
# fixture projects in tests/fixtures/, with absolute paths normalized to the
# literal placeholder <repo> so the output is stable across machines. The
# examples therefore cannot drift from the tool's actual behavior.
#
# The file is fully generated — do not edit docs/examples.md by hand. Change
# this script and run:  make docs-examples
# CI verifies idempotence with:  make docs-examples && git diff --exit-code
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

make build >/dev/null 2>&1
BIN="$ROOT/bin/repoctx"
FIX="tests/fixtures"
FIX_ABS="$ROOT/tests/fixtures"

# norm replaces the absolute fixtures root with the <repo> placeholder.
norm() {
	sed "s|$FIX_ABS|<repo>|g"
}

# run <args...> executes repoctx and prints its normalized stdout.
run() {
	"$BIN" "$@" 2>/dev/null | norm
}

render() {
	cat <<'EOF'
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

EOF

	printf '```console\n$ repoctx %s\n' "info tests/fixtures/scanner/mono"
	run info "$FIX/scanner/mono"
	printf '```\n\n'

	cat <<'EOF'
Same project, JSON output (the `root` field and `manifest[].path` are part of
the documented contract):

EOF

	printf '```json\n$ repoctx info tests/fixtures/scanner/mono --json\n'
	run info "$FIX/scanner/mono" --json
	printf '```\n\n'

	cat <<'EOF'
A corrupt manifest is a soft failure: the scan keeps the broken manifest with
zero facts and reports why (`!` in human output, `errors` in JSON), and the
exit code stays `0`. Here `broken/package.json` is not valid JSON:

EOF

	printf '```console\n$ repoctx %s\n' "info tests/fixtures/scanner/corrupt"
	run info "$FIX/scanner/corrupt"
	printf '```\n\n'

	printf '```json\n$ repoctx info tests/fixtures/scanner/corrupt --json\n'
	run info "$FIX/scanner/corrupt" --json
	printf '```\n\n'

	cat <<'EOF'
## generate

`generate` rewrites only the content between the `<!-- repoctx:start -->` and
`<!-- repoctx:end -->` markers, preserving human-written prose. `--dry-run`
previews what would change without writing anything, and exits `0` either way.

Here the `ghost` fixture claims a `lint` script that was removed from its
`package.json`, so a generation would rewrite its Commands table:

EOF

	printf '```console\n$ repoctx %s\n' "generate tests/fixtures/audit/ghost --dry-run"
	run generate "$FIX/audit/ghost" --dry-run
	printf '```\n\n'

	cat <<'EOF'
Once a context file is fully in sync, the same dry-run reports it as current:

EOF

	printf '```console\n$ repoctx %s\n' "generate <fresh copy of healthy> --dry-run"
	FRESH="$(mktemp -d)"
	cp -r "$FIX/audit/healthy/." "$FRESH/"
	run generate "$FRESH" >/dev/null
	"$BIN" generate "$FRESH" --dry-run 2>/dev/null | norm
	printf '```\n\n'
	rm -rf "$FRESH"

	cat <<'EOF'
## audit

`audit` checks a context file's claims against the current code and scores the
drift. A fully in-sync file passes 100/100:

EOF

	printf '```console\n$ repoctx %s\n' "audit tests/fixtures/audit/healthy"
	run audit "$FIX/audit/healthy"
	printf '```\n\n'

	cat <<'EOF'
Rot is reported per check — a stale path here, a ghost command there:

EOF

	printf '```console\n$ repoctx %s\n' "audit tests/fixtures/audit/stale"
	run audit "$FIX/audit/stale"
	printf '```\n\n'

	printf '```console\n$ repoctx %s\n' "audit tests/fixtures/audit/ghost"
	run audit "$FIX/audit/ghost"
	printf '```\n\n'

	cat <<'EOF'
JSON output, one report per context file:

EOF

	printf '```json\n$ repoctx audit tests/fixtures/audit/healthy --json\n'
	run audit "$FIX/audit/healthy" --json
	printf '```\n\n'

	cat <<'EOF'
## Exit codes

`0` means success, `1` means any failure. `audit --check` turns rot into an
exit code so CI can gate merges — healthy passes, rotting context fails:

EOF

	printf '```console\n$ repoctx audit tests/fixtures/audit/healthy --check; echo $?\n'
	"$BIN" audit "$FIX/audit/healthy" --check 2>/dev/null | norm
	printf '%s\n```\n\n' "$?"

	printf '```console\n$ repoctx audit tests/fixtures/audit/ghost --check; echo $?\n'
	set +e
	GHOST_OUT="$("$BIN" audit "$FIX/audit/ghost" --check 2>/dev/null | norm)"
	GHOST_CODE=$?
	set -e
	printf '%s\n%s\n```\n\n' "$GHOST_OUT" "$GHOST_CODE"

	cat <<'EOF'
Invalid usage is a failure too:

EOF

	printf '```console\n$ repoctx info tests/fixtures/scanner/mono --max-depth 0; echo $?\n'
	set +e
	DEPTH_OUT="$("$BIN" info "$FIX/scanner/mono" --max-depth 0 2>&1 | norm)"
	DEPTH_CODE=$?
	set -e
	printf '%s\n%s\n```\n\n' "$DEPTH_OUT" "$DEPTH_CODE"

	cat <<'EOF'
## Scan tuning flags

`--max-depth` limits how far the scanner descends; `--skip-dirs` (repeatable)
adds directories to skip on top of the built-ins (`.git`, `node_modules`,
`vendor`, ...). The same flags work on `info`, `generate` and `audit`.

`--max-depth 1` on the monorepo keeps only the root manifest:

EOF

	printf '```console\n$ repoctx %s\n' "info tests/fixtures/scanner/mono --max-depth 1"
	run info "$FIX/scanner/mono" --max-depth 1
	printf '```\n\n'

	cat <<'EOF'
`--skip-dirs tools` drops the Rust tool directory:

EOF

	printf '```console\n$ repoctx %s\n' "info tests/fixtures/scanner/mono --skip-dirs tools"
	run info "$FIX/scanner/mono" --skip-dirs tools
	printf '```\n\n'

	cat <<'EOF'
The same tuning lives declaratively in a `repoctx.toml` in the scanned
directory, or in any location via `--config`. Precedence is
`flags > config > defaults`:

EOF

	printf '```console\n$ repoctx %s\n' "info tests/fixtures/scanner/mono --config <path to repoctx.toml>"
	CFG_DIR="$(mktemp -d)"
	printf 'skip_dirs = ["tools"]\n' >"$CFG_DIR/repoctx.toml"
	run info "$FIX/scanner/mono" --config "$CFG_DIR/repoctx.toml"
	printf '```\n\n'
	rm -rf "$CFG_DIR"

	cat <<'EOF'
## workflow

`workflow` prints a paste-ready block for the human-written section of a
context file, telling agents to regenerate and gate on truth:

EOF

	printf '```console\n$ repoctx workflow\n'
	run workflow
	printf '```\n\n'

	cat <<'EOF'
## Workspaces

Workspace (monorepo) layouts are scanned naturally: every manifest in the tree
is detected, commands are attributed per manifest, and identical commands from
different packages are disambiguated by the Source column.

npm/pnpm workspace with a root `package.json` and two member packages:

EOF

	printf '```json\n$ repoctx info tests/fixtures/scanner/workspace-npm --json\n'
	run info "$FIX/scanner/workspace-npm" --json
	printf '```\n\n'

	cat <<'EOF'
Cargo workspace: a virtual root (no `[package]`) exposes the workspace-wide
commands but not `cargo run`, which has no default binary; member crates keep
the full set. `[workspace.dependencies]` are extracted too:

EOF

	printf '```json\n$ repoctx info tests/fixtures/scanner/workspace-cargo --json\n'
	run info "$FIX/scanner/workspace-cargo" --json
	printf '```\n\n'

	cat <<'EOF'
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
EOF
}

render >"$ROOT/docs/examples.md"
echo "docs/examples.md regenerated"
