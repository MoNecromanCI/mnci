#!/bin/sh
# Install mnci from its GitHub Release.
#
#   curl -fsSL https://raw.githubusercontent.com/MoNecromanCI/mnci/main/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- 0.1.0            # a specific version
#
# Installs to $MNCI_INSTALL_DIR (default ~/.local/bin). On GitHub Actions and
# Azure Pipelines the directory is added to PATH for the following steps.
set -eu

REPO="MoNecromanCI/mnci"
VERSION="${1:-${MNCI_VERSION:-latest}}"
DIR="${MNCI_INSTALL_DIR:-$HOME/.local/bin}"

fail() { echo "mnci install: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "$1 is required"; }
need curl
need unzip

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  MINGW*|MSYS*|CYGWIN*) os=windows ;;
  *) fail "unsupported OS $(uname -s)" ;;
esac
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) fail "unsupported architecture $(uname -m)" ;;
esac

if [ "$VERSION" = "latest" ]; then
  # GitHub limits anonymous API calls per address, which shared CI runners
  # exhaust; a token (CI provides one) lifts the limit.
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    set -- -H "Authorization: Bearer $GITHUB_TOKEN"
  else
    set --
  fi
  # GitHub lists releases by tag name as text (mnci@0.9.0 before mnci@0.20.0),
  # so take the highest version, comparing its numbers.
  VERSION=$(curl -fsSL "$@" "https://api.github.com/repos/$REPO/releases?per_page=100" \
    | grep -o '"tag_name": *"mnci@[0-9][0-9.]*"' | sed 's/.*mnci@\([^"]*\)".*/\1/' \
    | sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1)
  [ -n "$VERSION" ] || fail "no mnci release found"
fi
VERSION="${VERSION#v}"

asset="mnci_${VERSION}_${os}_${arch}.zip"
url="https://github.com/$REPO/releases/download/mnci%40${VERSION}/${asset}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Installing mnci $VERSION ($os/$arch) into $DIR"
curl -fsSL -o "$tmp/$asset" "$url" || fail "could not download $url"
if curl -fsSL -o "$tmp/checksums.txt" "https://github.com/$REPO/releases/download/mnci%40${VERSION}/checksums.txt" 2>/dev/null; then
  expected=$(grep " $asset\$" "$tmp/checksums.txt" | cut -d' ' -f1)
  if [ -n "$expected" ]; then
    if command -v sha256sum >/dev/null 2>&1; then actual=$(sha256sum "$tmp/$asset" | cut -d' ' -f1)
    else actual=$(shasum -a 256 "$tmp/$asset" | cut -d' ' -f1); fi
    [ "$expected" = "$actual" ] || fail "checksum mismatch for $asset"
  fi
fi
unzip -q -o "$tmp/$asset" -d "$tmp/out"
mkdir -p "$DIR"
binary=mnci
[ "$os" = windows ] && binary=mnci.exe
mv "$tmp/out/$binary" "$DIR/$binary"
chmod +x "$DIR/$binary"

[ -n "${GITHUB_PATH:-}" ] && echo "$DIR" >> "$GITHUB_PATH"
[ -n "${TF_BUILD:-}" ] && echo "##vso[task.prependpath]$DIR"
case ":$PATH:" in *":$DIR:"*) ;; *) echo "Add $DIR to your PATH to use mnci." ;; esac
"$DIR/$binary" --version
