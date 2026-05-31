#!/bin/sh
set -eu

REPO_URL="https://github.com/vandordev/vx/releases"

require_cmd() {
	name=$1
	if ! command -v "$name" >/dev/null 2>&1; then
		printf 'vx installer requires %s\n' "$name" >&2
		exit 1
	fi
}

require_cmd uname
require_cmd mktemp
require_cmd curl
require_cmd tar

TMP_DIR=$(mktemp -d)

cleanup() {
	rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

VERSION=${VERSION:-}
if [ -n "$VERSION" ]; then
	RELEASE_URL="${REPO_URL}/download/${VERSION}"
	REQUESTED_VERSION="$VERSION"
else
	RELEASE_URL="${REPO_URL}/latest/download"
	REQUESTED_VERSION="latest"
fi

BIN_DIR=${BIN_DIR:-}
if [ -z "$BIN_DIR" ]; then
	if [ -z "${HOME:-}" ]; then
		printf 'vx installer requires HOME when BIN_DIR is unset\n' >&2
		exit 1
	fi
	BIN_DIR="$HOME/.local/bin"
fi

OS=$(uname -s)
ARCH=$(uname -m)

case "$OS/$ARCH" in
	Linux/x86_64)
		ASSET="vx-linux-amd64.tar.gz"
		EXTRACTED_BINARY="vx-linux-amd64"
		;;
	Linux/aarch64|Linux/arm64)
		ASSET="vx-linux-arm64.tar.gz"
		EXTRACTED_BINARY="vx-linux-arm64"
		;;
	Darwin/x86_64)
		ASSET="vx-darwin-amd64.tar.gz"
		EXTRACTED_BINARY="vx-darwin-amd64"
		;;
	Darwin/arm64)
		ASSET="vx-darwin-arm64.tar.gz"
		EXTRACTED_BINARY="vx-darwin-arm64"
		;;
	*)
		printf 'unsupported platform: %s/%s\n' "$OS" "$ARCH" >&2
		exit 1
		;;
esac

ARCHIVE_PATH="$TMP_DIR/$ASSET"
DOWNLOAD_URL="$RELEASE_URL/$ASSET"

if ! curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE_PATH"; then
	printf 'vx installer failed to download %s\n' "$DOWNLOAD_URL" >&2
	exit 1
fi

if ! tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"; then
	printf 'vx installer failed to extract %s\n' "$ARCHIVE_PATH" >&2
	exit 1
fi

if ! mkdir -p "$BIN_DIR"; then
	printf 'vx installer failed to create %s\n' "$BIN_DIR" >&2
	exit 1
fi

if ! mv "$TMP_DIR/$EXTRACTED_BINARY" "$BIN_DIR/vx"; then
	printf 'vx installer failed to install vx into %s\n' "$BIN_DIR" >&2
	exit 1
fi

if ! chmod +x "$BIN_DIR/vx"; then
	printf 'vx installer failed to mark %s/vx as executable\n' "$BIN_DIR" >&2
	exit 1
fi

printf 'installed vx %s to %s/vx\n' "$REQUESTED_VERSION" "$BIN_DIR"
printf 'verify with: vx --version\n'

case ":${PATH:-}:" in
	*:"$BIN_DIR":*)
		;;
	*)
		printf 'add to PATH: export PATH="%s:$PATH"\n' "$BIN_DIR"
		;;
esac
