#!/usr/bin/env bash
#
# Builds the marketplace package for the free repoctx skill: a ZIP with a
# single top-level folder "repoctx/" holding SKILL.md. That layout is what
# skill marketplaces expect — the folder name is the skill name, and the
# frontmatter inside SKILL.md has to agree with it.
#
# Usage: bash scripts/package-skill.sh
#
# Output: dist/repoctx-skill.zip
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

SKILL_DIR="skills/repoctx"
SKILL_NAME="$(basename "$SKILL_DIR")"
SKILL_MD="$SKILL_DIR/SKILL.md"
STAGE="dist/pkg/$SKILL_NAME"
OUT="dist/repoctx-skill.zip"

if [ ! -f "$SKILL_MD" ]; then
	echo "missing $SKILL_MD" >&2
	exit 1
fi

# The frontmatter name and the folder name must agree, otherwise the
# marketplace keys the listing by one name and installs another.
fm_name="$(sed -n 's/^name:[[:space:]]*//p' "$SKILL_MD" | head -1)"
if [ "$fm_name" != "$SKILL_NAME" ]; then
	echo "frontmatter name '$fm_name' does not match folder '$SKILL_NAME'" >&2
	exit 1
fi

fm_description="$(sed -n 's/^description:[[:space:]]*//p' "$SKILL_MD" | head -1)"
if [ -z "$fm_description" ]; then
	echo "missing 'description' in $SKILL_MD frontmatter" >&2
	exit 1
fi

# Publishing a skill that drifted from the CLI output is the one failure a
# buyer cannot recover from, so check it when a binary is at hand.
BIN=""
if [ -x "bin/repoctx" ]; then
	BIN="bin/repoctx"
elif command -v repoctx >/dev/null 2>&1; then
	BIN="$(command -v repoctx)"
fi
if [ -n "$BIN" ]; then
	tmp="$(mktemp)"
	"$BIN" skill >"$tmp"
	if ! diff -q "$tmp" "$SKILL_MD" >/dev/null; then
		echo "$SKILL_MD drifted from '$BIN skill' — run 'repoctx skill' and commit" >&2
		rm -f "$tmp"
		exit 1
	fi
	rm -f "$tmp"
else
	echo "note: no repoctx binary found, skipping the skill drift check"
fi

rm -rf dist/pkg
mkdir -p "$STAGE"
cp "$SKILL_MD" "$STAGE/SKILL.md"

# The marketplace expects more than a bare SKILL.md: a README in the skill
# folder and at least one supporting file. Those live in marketplace/ so the
# canonical skill stays the single file every agent channel installs.
if [ -d "marketplace/$SKILL_NAME" ]; then
	cp -R "marketplace/$SKILL_NAME/." "$STAGE/"
fi

supporting="$(find "$STAGE" -type f ! -name SKILL.md | wc -l)"
if [ "$supporting" -lt 1 ]; then
	echo "the package has no supporting file — the marketplace expects at least one" >&2
	exit 1
fi

rm -f "$OUT"

python3 - "$OUT" "$STAGE" << 'PY'
import os, sys, zipfile

out, stage = sys.argv[1], sys.argv[2]
top = os.path.dirname(stage)
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for root, dirs, files in os.walk(stage):
        dirs.sort()
        for f in sorted(files):
            p = os.path.join(root, f)
            z.write(p, os.path.relpath(p, top))
PY

rm -rf dist/pkg

echo "built: $OUT"
echo
echo "contents:"
python3 - "$OUT" << 'PY'
import sys, zipfile
for n in sorted(zipfile.ZipFile(sys.argv[1]).namelist()):
    print(n)
PY
echo
echo "next: upload $OUT to the listing in the marketplace creator dashboard"
