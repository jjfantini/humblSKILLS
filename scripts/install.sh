#!/bin/sh
# humblskills installer. Fetches the latest (or pinned) release archive from
# GitHub, verifies its SHA-256 against the published checksums.txt, and drops
# the binary onto $PATH. POSIX sh; no bash-isms.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | sh
#   curl -fsSL .../install.sh | VERSION=0.1.0 sh
#   curl -fsSL .../install.sh | INSTALL_DIR=$HOME/.local/bin sh

set -eu

REPO="jjfantini/humblSKILLS"
BIN="humblskills"
PREFIX="${PREFIX:-/usr/local}"
INSTALL_DIR="${INSTALL_DIR:-$PREFIX/bin}"
VERSION="${VERSION:-}"

log()  { printf '%s\n' "$*" >&2; }
die()  { log "error: $*"; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

have curl || die "curl is required"
have tar  || die "tar is required"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux)  os_tag=linux ;;
  darwin) os_tag=macos ;;
  *) die "unsupported OS: $os — download a windows/* archive manually from the releases page" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  arch_tag=amd64 ;;
  arm64|aarch64) arch_tag=arm64 ;;
  *) die "unsupported arch: $arch" ;;
esac

# Resolve VERSION from /releases/latest (plain v* tags, not the sibling
# cli/v* tag that only exists for go install).
if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | head -n1 \
    | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"v([^"]+)".*/\1/')
fi
[ -n "$VERSION" ] || die "could not discover a release — pass VERSION=X.Y.Z"

tag="v${VERSION}"
archive="${BIN}_${VERSION}_${os_tag}_${arch_tag}.tar.gz"
base="https://github.com/${REPO}/releases/download/${tag}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

log "downloading ${archive}"
curl -fsSL "${base}/${archive}"     -o "${tmp}/${archive}"     || die "download failed: ${base}/${archive}"
curl -fsSL "${base}/checksums.txt"  -o "${tmp}/checksums.txt"  || die "checksums download failed"

# Verify. Prefer shasum on macOS, sha256sum on linux.
want=$(grep "  ${archive}\$" "${tmp}/checksums.txt" | awk '{print $1}')
[ -n "$want" ] || die "no checksum for ${archive} in checksums.txt"

if have sha256sum; then
  got=$(sha256sum "${tmp}/${archive}" | awk '{print $1}')
elif have shasum; then
  got=$(shasum -a 256 "${tmp}/${archive}" | awk '{print $1}')
else
  die "need sha256sum or shasum to verify checksum"
fi
[ "$want" = "$got" ] || die "checksum mismatch: want $want got $got"
log "checksum verified"

tar -xzf "${tmp}/${archive}" -C "${tmp}"

if [ ! -d "$INSTALL_DIR" ]; then
  log "creating $INSTALL_DIR"
  mkdir -p "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "${tmp}/${BIN}" "${INSTALL_DIR}/${BIN}"
else
  log "escalating to write to ${INSTALL_DIR}"
  sudo install -m 0755 "${tmp}/${BIN}" "${INSTALL_DIR}/${BIN}"
fi

log ""
log "installed ${BIN} ${VERSION} → ${INSTALL_DIR}/${BIN}"
if ! printf '%s' ":$PATH:" | grep -q ":${INSTALL_DIR}:"; then
  log "note: ${INSTALL_DIR} is not on your PATH — add it to your shell init"
fi
log "run: ${BIN} doctor"
