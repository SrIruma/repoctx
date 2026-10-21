#!/usr/bin/env bash
#
# Installs the repoctx binary from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/SrIruma/repoctx/main/scripts/install.sh | sh
#
# The script detects the platform, downloads the matching binary and the
# release's SHA256SUMS.txt, verifies the checksum, and installs the binary. It
# fails closed: a checksum mismatch aborts before anything is installed.
#
# Options:
#   --version vX.Y.Z   install a specific release (default: latest)
#   --bindir DIR       install directory (default: $HOME/.local/bin)
#   --repo OWNER/NAME  source repository (default: SrIruma/repoctx)
#   -h, --help         show this help
set -euo pipefail

REPO="SrIruma/repoctx"
VERSION=""
BINDIR="${BINDIR:-${HOME}/.local/bin}"

usage() {
	cat <<'EOF'
Installs the repoctx binary from a GitHub release.

Usage: install.sh [--version vX.Y.Z] [--bindir DIR] [--repo OWNER/NAME]

  --version vX.Y.Z   install a specific release (default: latest)
  --bindir DIR       install directory (default: $HOME/.local/bin)
  --repo OWNER/NAME  source repository (default: SrIruma/repoctx)
  -h, --help         show this help
EOF
}

die() {
	echo "install: $*" >&2
	exit 1
}

while [ $# -gt 0 ]; do
	case "$1" in
	--version)
		[ $# -ge 2 ] || die "--version needs a value"
		VERSION="$2"
		shift 2
		;;
	--bindir)
		[ $# -ge 2 ] || die "--bindir needs a value"
		BINDIR="$2"
		shift 2
		;;
	--repo)
		[ $# -ge 2 ] || die "--repo needs a value"
		REPO="$2"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		die "unknown option: $1 (try --help)"
		;;
	esac
done

# --- platform detection -----------------------------------------------------

os="$(uname -s)"
case "$os" in
Linux) os="linux" ;;
Darwin) os="darwin" ;;
*) die "unsupported operating system: $os" ;;
esac

arch="$(uname -m)"
case "$arch" in
x86_64 | amd64) arch="amd64" ;;
arm64 | aarch64) arch="arm64" ;;
*) die "unsupported architecture: $arch" ;;
esac

asset="repoctx_${os}_${arch}"
[ "$os" = "windows" ] && asset="${asset}.exe"

# --- download ---------------------------------------------------------------

if [ -n "$VERSION" ]; then
	base="https://github.com/${REPO}/releases/download/${VERSION}"
else
	base="https://github.com/${REPO}/releases/latest/download"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "install: downloading ${asset} from ${base}"
curl -fsSL "${base}/${asset}" -o "${tmp}/${asset}" ||
	die "failed to download ${base}/${asset}"

curl -fsSL "${base}/SHA256SUMS.txt" -o "${tmp}/SHA256SUMS.txt" ||
	die "failed to download SHA256SUMS.txt (cannot verify integrity)"

# --- verify -----------------------------------------------------------------

expected="$(grep " ${asset}\$" "${tmp}/SHA256SUMS.txt" | awk '{print $1}')"
[ -n "$expected" ] || die "no checksum for ${asset} in SHA256SUMS.txt"

if command -v sha256sum >/dev/null 2>&1; then
	actual="$(sha256sum "${tmp}/${asset}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
	actual="$(shasum -a 256 "${tmp}/${asset}" | awk '{print $1}')"
else
	die "neither sha256sum nor shasum is available to verify the download"
fi

[ "$expected" = "$actual" ] || die "checksum mismatch for ${asset}
  expected: ${expected}
  actual:   ${actual}"

echo "install: checksum verified"

# --- install ----------------------------------------------------------------

mkdir -p "$BINDIR"
chmod +x "${tmp}/${asset}"
mv "${tmp}/${asset}" "${BINDIR}/repoctx"
echo "install: repoctx installed to ${BINDIR}/repoctx"

case ":${PATH}:" in
*":${BINDIR}:"*) ;;
*) echo "install: add ${BINDIR} to your PATH to run repoctx" ;;
esac
