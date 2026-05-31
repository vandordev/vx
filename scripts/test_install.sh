#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

pass() {
	printf 'ok - %s\n' "$1"
}

fail() {
	printf 'not ok - %s\n' "$1" >&2
	exit 1
}

assert_contains() {
	needle=$1
	file=$2
	if ! grep -F "$needle" "$file" >/dev/null 2>&1; then
		printf 'expected to find: %s\n' "$needle" >&2
		printf 'in output:\n' >&2
		cat "$file" >&2
		exit 1
	fi
}

make_stub_dir() {
	workdir=$1
	mkdir -p "$workdir/stubs"

	cat >"$workdir/stubs/uname" <<'EOF'
#!/usr/bin/env sh
case "${1:-}" in
	-s) printf '%s\n' "${VX_TEST_UNAME_S:-Linux}" ;;
	-m) printf '%s\n' "${VX_TEST_UNAME_M:-x86_64}" ;;
	*) exit 1 ;;
esac
EOF

	cat >"$workdir/stubs/curl" <<'EOF'
#!/usr/bin/env sh
out=''
while [ "$#" -gt 0 ]; do
	case "$1" in
		-o)
			out=$2
			shift 2
			;;
		*)
			last=$1
			shift
			;;
	esac
done
	printf '%s\n' "${last:-}" >"${VX_TEST_CURL_LOG}"
	: >"$out"
EOF

	cat >"$workdir/stubs/tar" <<'EOF'
#!/usr/bin/env sh
dest=''
while [ "$#" -gt 0 ]; do
	case "$1" in
		-C)
			dest=$2
			shift 2
			;;
		*)
			shift
			;;
	esac
done
	: "${dest:?missing tar destination}"
	printf 'binary\n' >"$dest/${VX_TEST_EXTRACTED_BINARY}"
EOF

	chmod +x "$workdir/stubs/uname" "$workdir/stubs/curl" "$workdir/stubs/tar"
}

run_installer() {
	workdir=$1
	output=$2
	shift 2

	PATH="$workdir/stubs:$PATH" \
	VX_TEST_CURL_LOG="$workdir/curl.log" \
	VX_TEST_EXTRACTED_BINARY="${VX_TEST_EXTRACTED_BINARY:-vx-linux-amd64}" \
	"$@" sh "$ROOT_DIR/scripts/install.sh" >"$output" 2>&1
}

test_latest_linux_default_dir() {
	workdir=$(mktemp -d)
	trap 'rm -rf "$workdir"' EXIT INT TERM
	make_stub_dir "$workdir"

	output="$workdir/output.log"
	home_dir="$workdir/home"
	mkdir -p "$home_dir"

	if ! run_installer "$workdir" "$output" env HOME="$home_dir"; then
		cat "$output" >&2
		fail "latest install on Linux/x86_64 succeeds"
	fi

	test -x "$home_dir/.local/bin/vx" || fail "latest install on Linux/x86_64 writes vx to default bin dir"
	assert_contains "https://github.com/vandordev/vx/releases/latest/download/vx-linux-amd64.tar.gz" "$workdir/curl.log"
	pass "latest install on Linux/x86_64 succeeds"
	rm -rf "$workdir"
	trap - EXIT INT TERM
}

test_pinned_darwin_custom_bin_dir() {
	workdir=$(mktemp -d)
	trap 'rm -rf "$workdir"' EXIT INT TERM
	make_stub_dir "$workdir"

	output="$workdir/output.log"
	custom_bin="$workdir/custom-bin"

	if ! run_installer "$workdir" "$output" env VX_TEST_UNAME_S=Darwin VX_TEST_UNAME_M=arm64 VX_TEST_EXTRACTED_BINARY=vx-darwin-arm64 VERSION=v9.9.9 BIN_DIR="$custom_bin"; then
		cat "$output" >&2
		fail "pinned install on Darwin/arm64 with BIN_DIR succeeds"
	fi

	test -x "$custom_bin/vx" || fail "pinned install on Darwin/arm64 writes vx to custom bin dir"
	assert_contains "https://github.com/vandordev/vx/releases/download/v9.9.9/vx-darwin-arm64.tar.gz" "$workdir/curl.log"
	pass "pinned install on Darwin/arm64 with BIN_DIR succeeds"
	rm -rf "$workdir"
	trap - EXIT INT TERM
}

test_unsupported_platform_fails() {
	workdir=$(mktemp -d)
	trap 'rm -rf "$workdir"' EXIT INT TERM
	make_stub_dir "$workdir"

	output="$workdir/output.log"
	if run_installer "$workdir" "$output" env VX_TEST_UNAME_S=Plan9 VX_TEST_UNAME_M=amd64 HOME="$workdir/home"; then
		fail "unsupported platform fails"
	fi

	assert_contains "unsupported platform" "$output"
	pass "unsupported platform fails"
	rm -rf "$workdir"
	trap - EXIT INT TERM
}

test_missing_home_fails() {
	workdir=$(mktemp -d)
	trap 'rm -rf "$workdir"' EXIT INT TERM
	make_stub_dir "$workdir"

	output="$workdir/output.log"
	if run_installer "$workdir" "$output" env -u HOME; then
		fail "missing HOME fails when BIN_DIR is unset"
	fi

	assert_contains "HOME" "$output"
	pass "missing HOME fails when BIN_DIR is unset"
	rm -rf "$workdir"
	trap - EXIT INT TERM
}

test_latest_linux_default_dir
test_pinned_darwin_custom_bin_dir
test_unsupported_platform_fails
test_missing_home_fails
