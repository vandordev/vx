# VX Install Script Distribution Design

## Summary

`vx` currently has an idiomatic Go CLI layout where the executable entrypoint lives in `cmd/vx`. That keeps the codebase structure clean, but it means the correct Go install path is `go install github.com/vandordev/vx/cmd/vx@latest`, not `go install github.com/vandordev/vx@latest`.

This design adds a shorter end-user installation path without changing the Go module layout. The primary install experience becomes a release-backed shell installer for macOS and Linux. The existing `go install` path remains documented as the developer and cross-platform fallback, including Windows.

## Goals

- Provide a short install command for macOS and Linux users.
- Keep the current `cmd/vx` executable layout unchanged.
- Use GitHub Releases as the source of truth for official binaries.
- Support default installation of the latest release.
- Support optional version pinning for reproducible installs.
- Default to a user-space install location that does not require `sudo`.
- Keep Windows supported through `go install`.

## Non-Goals

- Making `go install github.com/vandordev/vx@latest` valid.
- Adding a Windows installer in this version.
- Automatically editing shell profile files.
- Requiring or assuming Homebrew for the primary install flow.
- Adding checksum or signature verification unless release artifacts already provide that surface.

## Primary User Experience

The primary documented installation path becomes:

```bash
curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

Optional variants:

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
BIN_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
VERSION=v0.1.0 BIN_DIR=$HOME/bin curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

The documented Go-based fallback remains:

```bash
go install github.com/vandordev/vx/cmd/vx@latest
```

## Platform Scope

Supported by `install.sh` in v1:

- macOS amd64
- macOS arm64
- Linux amd64
- Linux arm64

Not supported by `install.sh` in v1:

- Windows
- unsupported or unknown architectures

Windows users will be directed to:

```bash
go install github.com/vandordev/vx/cmd/vx@latest
```

## Release Artifact Contract

The installer depends on the existing release artifact naming produced by the repository release flow.

Expected archive names:

- `vx-darwin-amd64.tar.gz`
- `vx-darwin-arm64.tar.gz`
- `vx-linux-amd64.tar.gz`
- `vx-linux-arm64.tar.gz`

Release URLs:

- latest: `https://github.com/vandordev/vx/releases/latest/download/<asset>`
- pinned: `https://github.com/vandordev/vx/releases/download/<version>/<asset>`

This makes GitHub Releases the source of truth for binary distribution while keeping the shell installer stateless.

## Installer Interface

The installer entrypoint is:

- `scripts/install.sh`

Supported environment variables:

- `VERSION`
  - optional
  - default: unset, meaning install the latest release
  - when set, must reference a release tag such as `v0.1.0`
- `BIN_DIR`
  - optional
  - default: `$HOME/.local/bin`
  - when set, overrides the final install location

The installer does not accept positional arguments in v1. Environment-variable configuration keeps the `curl | sh` interface simple.

## Runtime Flow

1. Validate required shell tools such as `curl` and `tar`.
2. Read `VERSION` and `BIN_DIR`.
3. Detect OS from `uname -s`.
4. Detect architecture from `uname -m`.
5. Map OS and architecture to a supported release asset name.
6. Resolve the download URL from the asset name and selected version.
7. Download the archive into a temporary directory.
8. Extract the `vx` binary from the archive.
9. Create `BIN_DIR` if needed.
10. Move the binary into `BIN_DIR`.
11. Set executable permissions.
12. Print a success summary, installed path, and version hint.
13. Print a PATH warning if `BIN_DIR` is not currently present in the shell PATH.

## OS and Architecture Mapping

The installer maps runtime values to release assets using the following rules:

- `Darwin` + `x86_64` -> `vx-darwin-amd64.tar.gz`
- `Darwin` + `arm64` -> `vx-darwin-arm64.tar.gz`
- `Linux` + `x86_64` -> `vx-linux-amd64.tar.gz`
- `Linux` + `aarch64` -> `vx-linux-arm64.tar.gz`
- `Linux` + `arm64` -> `vx-linux-arm64.tar.gz`

All other combinations fail fast with a clear unsupported-platform error.

## Installation Target Behavior

Default target:

- `$HOME/.local/bin`

Override target:

- `BIN_DIR=/custom/path`

Behavioral constraints:

- The installer must not invoke `sudo` automatically.
- The installer must not silently fall back to another target directory.
- If `BIN_DIR` cannot be created or written, the installer exits non-zero with a direct explanation.

The choice of `$HOME/.local/bin` keeps the default flow safe for `curl | sh` by avoiding privilege escalation.

## Error Handling

The installer must fail early and clearly in these cases:

- missing `curl`
- missing `tar`
- unsupported OS
- unsupported architecture
- release asset not found
- download failure
- extraction failure
- unwritable install directory
- missing `HOME` when defaulting `BIN_DIR`

Error messages should be concrete and operational. They should name the missing dependency, unsupported platform, or expected release URL pattern where helpful.

## Success Output

On success, the installer should print:

- the installed binary path
- whether `latest` or a pinned `VERSION` was requested
- a `vx --version` verification hint

If `BIN_DIR` is not on the current PATH, the installer should also print a shell snippet the user can add manually. The installer must not mutate shell rc files automatically.

## Documentation Changes

The install surface should be updated in these files:

- `README.md`
- `INSTALL.md`
- generated docs under `docs/src/content/docs/`

Documentation priorities:

1. show `curl -fsSL ... | sh` as the primary install path
2. show `VERSION=...` and `BIN_DIR=...` variants
3. keep `go install github.com/vandordev/vx/cmd/vx@latest` as the developer and Windows fallback
4. make clear why `go install github.com/vandordev/vx@latest` is not valid for this repo layout

## Verification Strategy

This work is partly shell distribution logic and partly documentation. The implementation should therefore verify both artifact compatibility and user-facing docs.

Minimum verification expectations:

- confirm the installer asset mapping matches the names produced by the release tooling
- run the installer in a controlled local mode or equivalent validation path
- verify generated docs were refreshed if source markdown changed
- verify shell script behavior for supported and unsupported platform branches where practical

Suggested concrete checks:

- inspect release packaging scripts and asset naming
- run `just docs-generate` after doc updates
- run installer checks with overridden environment where possible

## Alternatives Considered

### Keep `go install` as the only public install path

Rejected because it does not improve the end-user experience and still exposes the longer `cmd/vx` path.

### Move the executable entrypoint to the module root

Rejected because it weakens the current Go project layout just to shorten one install command.

### Make Homebrew the primary install path

Rejected for now because it improves macOS UX but does not solve Linux distribution symmetry and adds higher release maintenance overhead.

## Recommended Approach

Implement `scripts/install.sh` as the primary macOS/Linux install path, backed by GitHub Releases and defaulting to `$HOME/.local/bin`. Keep `go install github.com/vandordev/vx/cmd/vx@latest` as the documented developer and Windows fallback. This improves installation UX without changing the current Go module structure or broadening runtime scope.
