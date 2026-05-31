# VX Install Script Distribution Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a release-backed `install.sh` for macOS and Linux so end users get a short install command, while keeping `go install github.com/vandordev/vx/cmd/vx@latest` as the developer and Windows fallback.

**Architecture:** Keep the existing Go module layout unchanged and implement distribution as a thin POSIX `sh` installer in `scripts/install.sh`. Test it black-box with a dedicated shell test harness that stubs `uname`, `curl`, `tar`, `mv`, and `chmod`, then update only the source docs that feed the generated docs site.

**Tech Stack:** POSIX `sh`, existing GitHub Releases artifact naming from `scripts/github_release.sh`, `just`, `go test`, shell test scripts, `just docs-generate`.

---

## File Structure

### Existing files to modify

- `justfile`
  - Add focused installer verification targets without changing the existing `just test` Go suite.
- `README.md`
  - Promote the short installer command as the primary install path.
- `INSTALL.md`
  - Document `curl | sh`, `VERSION`, `BIN_DIR`, and the Windows `go install` fallback.

### New files to create

- `scripts/install.sh`
  - POSIX `sh` installer for macOS and Linux release binaries.
- `scripts/test_install.sh`
  - Black-box shell tests for installer behavior using stubbed system tools.
- `scripts/test_install_docs.sh`
  - Small docs assertions to keep install instructions aligned with the supported distribution surface.

### Generated files expected to change after docs regeneration

- `docs/src/content/docs/index.md`
- `docs/src/content/docs/install.md`

These are generated outputs. Do not edit them directly; regenerate them from `README.md` and `INSTALL.md`.

## Chunk 1: Installer Contract And Core Happy Path

### Task 1: Add a failing black-box test harness for the installer

**Files:**
- Create: `scripts/test_install.sh`
- Modify: `justfile`

- [ ] **Step 1: Write failing installer tests for the supported-path contract**

Create tests that cover:

- latest install on `Linux/x86_64` to default `$HOME/.local/bin`
- pinned install on `Darwin/arm64` with `BIN_DIR` override
- unsupported platform failure for an unknown `uname -s`
- missing `HOME` failure when `BIN_DIR` is unset

Use stub executables in a temp `PATH` so the tests never hit the network or the real filesystem outside the temp workspace.

- [ ] **Step 2: Add a focused just target for the installer tests**

Add:

```just
test-install:
	./scripts/test_install.sh
```

- [ ] **Step 3: Run the installer tests to verify they fail**

Run: `just test-install`

Expected: FAIL because `scripts/install.sh` does not exist yet.

- [ ] **Step 4: Commit the failing installer test baseline**

```bash
git add justfile scripts/test_install.sh
git commit -m "test: add install script harness"
```

### Task 2: Implement the minimal installer flow to satisfy the core contract

**Files:**
- Create: `scripts/install.sh`

- [ ] **Step 1: Implement a POSIX `sh` script skeleton with strict mode and cleanup**

Start with:

```sh
#!/bin/sh
set -eu
```

Use a temp directory plus a cleanup trap. Keep the script POSIX-compatible because the documented interface is `curl ... | sh`.

- [ ] **Step 2: Implement version and bin-dir resolution**

Rules:

- `VERSION` unset -> latest release URL form
- `VERSION=v0.1.0` -> pinned release URL form
- `BIN_DIR` unset -> `$HOME/.local/bin`
- if `HOME` is missing and `BIN_DIR` is unset -> fail fast

- [ ] **Step 3: Implement OS and architecture mapping**

Map:

- `Darwin` + `x86_64` -> `vx-darwin-amd64.tar.gz`
- `Darwin` + `arm64` -> `vx-darwin-arm64.tar.gz`
- `Linux` + `x86_64` -> `vx-linux-amd64.tar.gz`
- `Linux` + `aarch64` -> `vx-linux-arm64.tar.gz`
- `Linux` + `arm64` -> `vx-linux-arm64.tar.gz`

All other combinations must exit non-zero with a clear unsupported-platform message.

- [ ] **Step 4: Implement the minimal download, extract, and install flow**

The script should:

1. build the GitHub Releases URL
2. download the archive to a temp file
3. extract the platform-named binary
4. install it to `${BIN_DIR}/vx`
5. `chmod +x` the final binary

Handle the fact that the archive contains a platform-named binary such as `vx-linux-amd64`, not a plain `vx`.

- [ ] **Step 5: Run the installer tests to verify they pass**

Run: `just test-install`

Expected: PASS for the happy-path and basic failure-path tests added in Task 1.

- [ ] **Step 6: Commit the core installer**

```bash
git add scripts/install.sh
git commit -m "feat: add core install script"
```

## Chunk 2: Error Handling And Operator Feedback

### Task 3: Extend the tests for dependency failures and PATH messaging

**Files:**
- Modify: `scripts/test_install.sh`

- [ ] **Step 1: Add a failing test for missing `curl`**

Simulate a `PATH` where `curl` is absent and assert the script exits non-zero with a concrete missing-dependency message.

- [ ] **Step 2: Add a failing test for missing `tar`**

Simulate a `PATH` where `tar` is absent and assert the script exits non-zero with a concrete missing-dependency message.

- [ ] **Step 3: Add a failing test for a write/install failure**

Stub `mv` or the target directory creation so the install step fails, and assert the script surfaces the write failure instead of silently succeeding.

- [ ] **Step 4: Add a failing test for the PATH warning**

Install successfully with a `BIN_DIR` that is not in `PATH` and assert the script prints a manual export hint.

- [ ] **Step 5: Run the installer tests to verify the new cases fail**

Run: `just test-install`

Expected: FAIL because the current script does not yet report all dependency and PATH cases correctly.

- [ ] **Step 6: Commit the expanded failing tests**

```bash
git add scripts/test_install.sh
git commit -m "test: cover install script failure paths"
```

### Task 4: Implement dependency checks and clearer operator messaging

**Files:**
- Modify: `scripts/install.sh`

- [ ] **Step 1: Add explicit command checks before runtime work**

Validate at least:

- `curl`
- `tar`
- `uname`
- `mktemp`

Print direct messages such as:

```text
vx installer requires curl
vx installer requires tar
```

- [ ] **Step 2: Tighten install-directory error handling**

Fail clearly when:

- the target directory cannot be created
- the extracted binary cannot be moved into place
- `chmod +x` fails

Do not invoke `sudo` automatically and do not silently change install destinations.

- [ ] **Step 3: Print a concise success summary**

Include:

- installed path
- whether the request was `latest` or the pinned `VERSION`
- a `vx --version` hint

- [ ] **Step 4: Print a PATH warning when needed**

If `BIN_DIR` is not present in `PATH`, print a manual export line such as:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

- [ ] **Step 5: Run the installer tests to verify they pass**

Run: `just test-install`

Expected: PASS for all installer test cases.

- [ ] **Step 6: Commit the hardened installer behavior**

```bash
git add scripts/install.sh scripts/test_install.sh
git commit -m "feat: harden install script behavior"
```

## Chunk 3: Documentation Surface And Full Verification

### Task 5: Add failing docs assertions for the public install surface

**Files:**
- Create: `scripts/test_install_docs.sh`
- Modify: `justfile`

- [ ] **Step 1: Write failing docs assertions for the primary install path**

Assert that:

- `README.md` promotes `curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh`
- `INSTALL.md` documents `VERSION=...` and `BIN_DIR=...`
- `INSTALL.md` keeps `go install github.com/vandordev/vx/cmd/vx@latest`
- `INSTALL.md` explicitly points Windows users to `go install`

- [ ] **Step 2: Add a focused just target for docs assertions**

Add:

```just
test-install-docs:
	./scripts/test_install_docs.sh
```

- [ ] **Step 3: Run the docs assertions to verify they fail**

Run: `just test-install-docs`

Expected: FAIL because the current docs still lead with `go install`.

- [ ] **Step 4: Commit the failing docs assertions**

```bash
git add justfile scripts/test_install_docs.sh
git commit -m "test: lock install documentation surface"
```

### Task 6: Update the source docs and regenerate generated docs

**Files:**
- Modify: `README.md`
- Modify: `INSTALL.md`
- Generate: `docs/src/content/docs/index.md`
- Generate: `docs/src/content/docs/install.md`

- [ ] **Step 1: Update `README.md` to promote the short installer command first**

Show:

```bash
curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

Keep the Go install path as the developer and Windows fallback.

- [ ] **Step 2: Update `INSTALL.md` to document the full install matrix**

Document:

- primary `curl | sh` install
- `VERSION=...`
- `BIN_DIR=...`
- Windows via `go install`
- why module-root `go install github.com/vandordev/vx@latest` is still invalid

- [ ] **Step 3: Run the docs assertions to verify they pass**

Run: `just test-install-docs`

Expected: PASS.

- [ ] **Step 4: Regenerate the docs site content**

Run: `just docs-generate`

Expected: PASS and update generated docs under `docs/src/content/docs/`.

- [ ] **Step 5: Commit the docs surface**

```bash
git add README.md INSTALL.md docs/src/content/docs/index.md docs/src/content/docs/install.md
git commit -m "docs: promote install script distribution"
```

### Task 7: Run the full verification set for the distribution feature

**Files:**
- Use: `scripts/install.sh`
- Use: `scripts/test_install.sh`
- Use: `scripts/test_install_docs.sh`
- Use: `scripts/github_release.sh`

- [ ] **Step 1: Verify the focused installer tests**

Run: `just test-install`

Expected: PASS.

- [ ] **Step 2: Verify the docs assertions**

Run: `just test-install-docs`

Expected: PASS.

- [ ] **Step 3: Verify the Go test suite still passes**

Run: `just test`

Expected: PASS.

- [ ] **Step 4: Verify docs regeneration is clean**

Run: `just docs-generate`

Expected: PASS with no unexpected diffs after regeneration.

- [ ] **Step 5: Verify the installer asset names still match the release script contract**

Run: `rg -n 'vx-(linux|darwin)-(amd64|arm64)(\\.tar\\.gz)?|windows/amd64' scripts/github_release.sh scripts/install.sh`

Expected: matching `linux` and `darwin` asset naming between installer logic and release tooling, with Windows remaining release-only and not installer-supported.

- [ ] **Step 6: Commit the verification checkpoint**

```bash
git status --short --branch
git commit --allow-empty -m "chore: verify install script distribution"
```
