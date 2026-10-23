#!/usr/bin/env bash
#
# Stages the built release binaries into the npm platform packages so they can
# be published. Run after `make release VERSION=...`.
#
# Usage: bash scripts/npm-stage.sh <version>
#   e.g. bash scripts/npm-stage.sh v0.7.0
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

version="${1:?usage: npm-stage.sh <version>}"
version="${version#v}"

# Map npm platform/arch -> Makefile artifact name.
stage() {
    local npmname="$1" artifact="$2"
    local dir="npm/repoctx-${npmname}"
    local bin="repoctx"
    mkdir -p "${dir}/bin"
    case "$npmname" in win32-*) bin="repoctx.exe" ;; esac
    if [ ! -f "dist/${artifact}" ]; then
        echo "missing dist/${artifact}" >&2
        exit 1
    fi
    cp "dist/${artifact}" "${dir}/bin/${bin}"
    cp npm/README.md "${dir}/README.md"
    chmod +x "${dir}/bin/${bin}" 2>/dev/null || true
    # keep package.json version in sync
    python3 - "$dir/package.json" "$version" << 'PY'
import json,sys
p=sys.argv[1]; v=sys.argv[2]
d=json.load(open(p)); d['version']=v; json.dump(d,open(p,'w'),indent=2); open(p,'a').write('\n')
PY
    echo "staged ${artifact} -> ${dir}/bin/${bin}"
}

stage darwin-arm64 repoctx_darwin_arm64
stage darwin-x64    repoctx_darwin_amd64
stage linux-arm64   repoctx_linux_arm64
stage linux-x64     repoctx_linux_amd64
stage win32-arm64   repoctx_windows_arm64.exe
stage win32-x64     repoctx_windows_amd64.exe

# main package version
mkdir -p npm/repoctx/bin
cp npm/README.md npm/repoctx/README.md
python3 - npm/repoctx/package.json "$version" << 'PY'
import json,sys
p=sys.argv[1]; v=sys.argv[2]
d=json.load(open(p)); d['version']=v
for k in d.get('optionalDependencies',{}):
    d['optionalDependencies'][k]=v
json.dump(d,open(p,'w'),indent=2); open(p,'a').write('\n')
PY
echo "staged main package @ ${version}"
