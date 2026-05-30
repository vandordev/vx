# VX Rebrand Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebrand the repository from `go-cli-template` to `vx`, preserve the full feature set, verify the repo under the new identity, then reset git history.

**Architecture:** Keep the existing CLI, docs, and release architecture in place, but replace all identity-bearing metadata and generated references so the codebase consistently builds and documents itself as `vx`. Perform the destructive git reset only after all text cleanup and runtime verification are complete.

**Tech Stack:** Go, Cobra, Bubble Tea, Lip Gloss, Bun, Astro Starlight, Just, shell scripts

---

## Chunk 1: Identity And Entrypoint

### Task 1: Rebrand core metadata and build targets

**Files:**
- Modify: `go.mod`
- Modify: `justfile`
- Modify: `internal/package/package.toml`
- Modify: `example-config.toml`

- [ ] **Step 1: Update metadata values to `vx`**

Set:

- module to `github.com/vandordev/vx`
- package name and binary name to `vx`
- repository URL to `https://github.com/vandordev/vx`

- [ ] **Step 2: Run targeted search to confirm the old identity still exists before code updates**

Run: `rg -n "go-cli-template|imdevan/go-cli-template" go.mod justfile internal/package example-config.toml -S`
Expected: matches present

- [ ] **Step 3: Update build/install targets**

Change binary output and install paths from `go-cli-template` to `vx`.

- [ ] **Step 4: Re-run search on the same files**

Run: `rg -n "go-cli-template|imdevan/go-cli-template" go.mod justfile internal/package example-config.toml -S`
Expected: no matches

### Task 2: Rename CLI entrypoint and fix imports

**Files:**
- Move: `cmd/go-cli-template` -> `cmd/vx`
- Modify: `cmd/vx/root.go`
- Modify: `cmd/vx/config.go`
- Modify: `cmd/vx/config_init.go`
- Modify: `cmd/vx/completion.go`
- Modify: `internal/config/interface.go`
- Modify: `internal/config/manager.go`
- Modify: `internal/config/manager_test.go`
- Modify: `internal/testutil/config.go`
- Modify: `internal/ui/theme.go`
- Modify: `internal/utils/paths.go`

- [ ] **Step 1: Update tests first for import-path expectations where needed**

Adjust any tests that import the old module path so they compile against `github.com/vandordev/vx`.

- [ ] **Step 2: Run a focused test or compile check and confirm failure**

Run: `go test ./internal/config ./internal/testutil ./internal/ui ./internal/utils`
Expected: fail due to stale imports or unresolved module path until implementation is complete

- [ ] **Step 3: Rename the command directory and update imports**

Ensure all package imports compile under `github.com/vandordev/vx` and completion examples print `vx`.

- [ ] **Step 4: Re-run focused tests**

Run: `go test ./internal/config ./internal/testutil ./internal/ui ./internal/utils`
Expected: pass

## Chunk 2: Docs And Distribution

### Task 3: Rebrand primary user docs

**Files:**
- Modify: `README.md`
- Modify: `INSTALL.md`
- Modify: `CONFIG.md`
- Modify: `CONTRIBUTING.md`
- Modify: `docs/README.md`

- [ ] **Step 1: Update user-facing naming and repository links**

Replace template wording with `vx` wording and remove “use this as a template” guidance.

- [ ] **Step 2: Run targeted search across these files**

Run: `rg -n "go-cli-template|imdevan/go-cli-template|template" README.md INSTALL.md CONFIG.md CONTRIBUTING.md docs/README.md -S`
Expected: only intentional generic words remain; no stale product identity

### Task 4: Rebrand docs app config and generated content

**Files:**
- Modify: `docs/config.mjs`
- Modify: `docs/sidebar.mjs`
- Modify: `docs/package.json`
- Modify: `docs/bun.lock`
- Delete: `docs/package-lock.json`
- Modify: `docs/src/content/docs/index.md`
- Modify: `docs/src/content/docs/install.md`
- Modify: `docs/src/content/docs/configuration.md`
- Modify: `docs/src/content/docs/contributing.md`
- Modify: `docs/src/content/docs/commands/completion.md`
- Modify: `docs/src/content/docs/commands/config.md`
- Modify: `docs/src/content/docs/commands/config-init.md`
- Move: `docs/src/content/docs/commands/go-cli-template.md` -> `docs/src/content/docs/commands/vx.md`
- Modify: `docs/src/content/docs/guides/configuration.md`
- Modify: `docs/src/content/docs/guides/installation.md`
- Modify: `docs/src/content/docs/guides/quickstart.md`
- Modify: `docs/src/styles/custom.css`
- Modify: `scripts/docs_generate.sh`

- [ ] **Step 1: Update docs metadata and special-case logic**

Replace project name, base path, GitHub URLs, and remove the `go-cli-template` production visibility exception in docs generation.

- [ ] **Step 2: Update generated markdown and command page references**

Ensure command docs refer to `vx`, `cmd/vx`, and the new repository URLs.

- [ ] **Step 3: Remove unsupported npm lockfile**

Delete `docs/package-lock.json` because docs package management is standardized on Bun in this repository.

- [ ] **Step 4: Re-run search across docs assets**

Run: `rg -n "go-cli-template|imdevan/go-cli-template" docs scripts -S`
Expected: no matches except design and plan docs under `docs/superpowers/`

### Task 5: Rebrand release and packaging surfaces

**Files:**
- Modify: `.github/workflows/*` as needed
- Modify: `scripts/*.sh` as needed
- Modify: `aur/PKGBUILD.template` and related packaging files as needed

- [ ] **Step 1: Search for stale identity in workflows and packaging**

Run: `rg -n "go-cli-template|imdevan/go-cli-template" .github scripts aur -S`
Expected: matches present before edits

- [ ] **Step 2: Replace stale package, binary, and repo references**

Preserve the existing workflows and packaging behavior, but point them to `vx`.

- [ ] **Step 3: Re-run the same search**

Run: `rg -n "go-cli-template|imdevan/go-cli-template" .github scripts aur -S`
Expected: no matches

## Chunk 3: Verification And Git Reset

### Task 6: End-to-end verification

**Files:**
- Verify only

- [ ] **Step 1: Run project tests through just**

Run: `just test`
Expected: pass

- [ ] **Step 2: Run direct Go test suite**

Run: `go test ./...`
Expected: pass

- [ ] **Step 3: Run build**

Run: `just build`
Expected: pass and produce `bin/vx`

- [ ] **Step 4: Run docs generation if environment supports it**

Run: `just docs-generate`
Expected: pass, or capture the exact environmental blocker if Bun or docs tooling is unavailable

- [ ] **Step 5: Run final stale-identity search**

Run: `rg -n "go-cli-template|imdevan/go-cli-template" . -S`
Expected: matches only inside `docs/superpowers/specs/` and `docs/superpowers/plans/`

### Task 7: Reset repository history

**Files:**
- Remove: `.git`
- Create: new `.git/`

- [ ] **Step 1: Confirm verification is complete**

Do not continue until Task 6 evidence is collected.

- [ ] **Step 2: Remove git history**

Run: `rm -rf .git`

- [ ] **Step 3: Initialize a fresh repository**

Run: `git init`
Expected: a new repository with no old commit history

- [ ] **Step 4: Confirm clean history state**

Run: `git log --oneline`
Expected: no commits yet
