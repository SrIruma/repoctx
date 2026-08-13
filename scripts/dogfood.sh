#!/usr/bin/env bash
#
# Dogfood rotation: run repoctx against real repositories and report what the
# scanner extracts and what would change in their context files.
#
# This is the feedback loop the repo's own dogfooding cannot provide: the
# repository here is regenerated in CI before it is audited, so it can never
# rot and never exercises the scanner against messy, real-world code. Point
# this at projects you actually work on.
#
# Usage:
#   scripts/dogfood.sh /path/to/repo [/path/to/another ...]
#
# Read-only: this script never writes to the target repositories.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

make build >/dev/null 2>&1
BIN="$ROOT/bin/repoctx"

for dir in "$@"; do
	if [ ! -d "$dir" ]; then
		echo "skip: $dir is not a directory" >&2
		continue
	fi
	echo "=== $dir ==="
	"$BIN" info "$dir" || true
	if [ -f "$dir/AGENTS.md" ] || [ -f "$dir/CLAUDE.md" ]; then
		"$BIN" audit "$dir" || true
		"$BIN" generate "$dir" --dry-run || true
	else
		echo "  (no AGENTS.md / CLAUDE.md; audit and generate skipped)"
	fi
	echo
done
