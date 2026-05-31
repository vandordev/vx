#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

assert_contains() {
	needle=$1
	file=$2
	if ! grep -F "$needle" "$file" >/dev/null 2>&1; then
		printf 'missing expected text in %s:\n%s\n' "$file" "$needle" >&2
		exit 1
	fi
}

README_FILE="$ROOT_DIR/README.md"
INSTALL_FILE="$ROOT_DIR/INSTALL.md"

assert_contains 'curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh' "$README_FILE"
assert_contains 'VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh' "$INSTALL_FILE"
assert_contains 'BIN_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh' "$INSTALL_FILE"
assert_contains 'go install github.com/vandordev/vx/cmd/vx@latest' "$INSTALL_FILE"
assert_contains 'Windows users should use:' "$INSTALL_FILE"

printf 'ok - install docs surface\n'
