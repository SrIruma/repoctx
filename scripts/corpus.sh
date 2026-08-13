#!/usr/bin/env bash
#
# External-corpus benchmark: clones real repositories at pinned SHAs, runs
# `repoctx info`, and diffs the extracted commands against the golden files in
# testdata/corpus/golden/.
#
# The golden files are the anti-circularity check. The repo's own dogfood CI
# only proves AGENTS.md matches what repoctx emits — that is consistency, not
# truth. The golden files here encode *expected* facts that a human reviewer
# checked against the pinned checkout, so a scanner regression that silently
# drops (or invents) a command shows up as a diff.
#
# Usage:
#   scripts/corpus.sh            # clone + compare, report, exit 0
#   scripts/corpus.sh --fail     # same, exit non-zero on any mismatch (CI)
#   scripts/corpus.sh --update    # write current output as the new golden
#
# Add a repository by appending `name\turl\tsha` to testdata/corpus/repos.tsv,
# then run `scripts/corpus.sh --update` and review the golden diff by hand.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MODE=check
case "${1:-}" in
	--fail) MODE=fail ;;
	--update) MODE=update ;;
esac

make build >/dev/null 2>&1
BIN="$ROOT/bin/repoctx"
CORPUS="$ROOT/testdata/corpus"
# Checkouts live outside the repository so they never leak into the repo's own
# `repoctx generate` / `audit` dogfooding (they are gitignored and would make
# the dogfood gate flap on a per-machine basis).
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/repoctx-corpus"
CHECKOUT="$CACHE_DIR/checkout"
ACTUAL="$(mktemp -d)"
trap 'rm -rf "$ACTUAL"' EXIT

mkdir -p "$CHECKOUT"

fail=0
while IFS=$'\t' read -r name url sha; do
	[ -n "$name" ] || continue
	[ "$name" != "name" ] || continue
	echo "== $name ($sha) =="
	dir="$CHECKOUT/$name"
	if [ ! -d "$dir/.git" ]; then
		git clone --depth 1 --quiet "$url" "$dir"
	fi
	git -C "$dir" fetch --depth 1 --quiet origin "$sha"
	git -C "$dir" checkout --quiet FETCH_HEAD

	"$BIN" info "$dir" --json >"$ACTUAL/$name.json"
	python3 - "$name" "$ACTUAL" <<'PY'
import json
import sys

name, actual = sys.argv[1], sys.argv[2]
with open(f"{actual}/{name}.json", encoding="utf-8") as fh:
    data = json.load(fh)
lines = []
for m in data.get("manifests") or []:
    for c in m.get("commands") or []:
        lines.append(f"{m['path']}\t{c['cmd']}")
lines.sort()
with open(f"{actual}/{name}.txt", "w", encoding="utf-8") as fh:
    fh.write("\n".join(lines) + ("\n" if lines else ""))
PY

	golden="$CORPUS/golden/$name.txt"
	case "$MODE" in
		update)
			cp "$ACTUAL/$name.txt" "$golden"
			echo "  golden updated"
			;;
		*)
			if diff -u "$golden" "$ACTUAL/$name.txt"; then
				echo "  golden match ($(wc -l <"$ACTUAL/$name.txt") commands)"
			else
				echo "  MISMATCH"
				fail=1
			fi
			;;
	esac
done <"$CORPUS/repos.tsv"

if [ "$MODE" = "update" ]; then
	echo "corpus: golden files updated — review the diff before committing"
	exit 0
fi
if [ "$fail" = "1" ]; then
	echo "corpus: FAIL (expected facts drift from what repoctx extracts)" >&2
	echo "  run scripts/corpus.sh, review the diff, and update the golden if the drift is real" >&2
	if [ "$MODE" = "fail" ]; then
		exit 1
	fi
fi
echo "corpus: done"
