#!/usr/bin/env bash
#
# Runs the same gates as .github/workflows/ci.yml locally, so a clean run here
# means CI will pass. Mirrors PROCESS.md §4 pre-flight exactly.
#
# Usage: bash scripts/preflight.sh
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

echo "== go vet =="
go vet ./...

echo "== gofmt =="
unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
    echo "$unformatted"
    exit 1
fi

echo "== go test =="
go test ./...

echo "== build =="
make build

echo "== generate drift (AGENTS.md) =="
./bin/repoctx generate . >/dev/null
git diff --exit-code AGENTS.md

echo "== audit =="
./bin/repoctx audit . --check

echo "== docs-examples drift =="
make docs-examples >/dev/null
git diff --exit-code docs/examples.md

echo "== corpus =="
bash scripts/corpus.sh --fail

echo "== preflight OK =="
