# VX Rebrand Design

## Goal

Transform the current repository from the upstream `go-cli-template` identity into a standalone `vx` CLI owned by `github.com/vandordev`, while preserving the existing feature set. Remove all legacy git history after the rebrand is complete and verified.

## Scope

This work keeps the current CLI architecture, Bubble Tea inline UI, configuration flow, docs generation, packaging scripts, release automation, and supporting tests.

This work replaces all template-specific identity and metadata, including:

- Go module path
- CLI command name and binary name
- import paths
- package metadata
- docs metadata and generated content inputs
- install and release references
- repository URLs and ownership strings
- config path names derived from project metadata

## Non-Goals

- Redesigning the product behavior of the CLI
- Removing the docs, packaging, or release toolchain
- Rewriting the architecture into a different framework
- Preserving any upstream template git history

## Recommended Approach

Use an in-place rebrand with a staged git reset at the end.

Rationale:

- It preserves the working feature surface with the lowest regression risk.
- It avoids reimplementing docs and release plumbing from scratch.
- It allows verification to happen before the repository history is destroyed.

## Planned Changes

## 1. Project identity

Update the repository to the following canonical values:

- CLI name: `vx`
- Go module: `github.com/vandordev/vx`
- Repository: `https://github.com/vandordev/vx`

All hardcoded `go-cli-template` and `imdevan/go-cli-template` references will be replaced or removed.

## 2. Entrypoint and imports

Rename `cmd/go-cli-template` to `cmd/vx` and update all build targets and import paths accordingly.

Expected outcomes:

- `go build ./cmd/vx` becomes the canonical build target
- generated docs and completion examples reference `vx`
- import paths compile under `github.com/vandordev/vx`

## 3. Metadata source of truth

Retain `internal/package/package.toml` as the single metadata source, but replace its template values with `vx` values so downstream sync and docs generation use the new identity.

Expected outcomes:

- package metadata helpers resolve `vx`
- config path helpers derive `~/.config/vx/config.toml`
- docs config and sidebar generation no longer include template exceptions

## 4. Documentation and distribution

Preserve the existing docs site and release scripts, but rebrand them fully:

- `README.md`, `INSTALL.md`, `CONFIG.md`, `CONTRIBUTING.md`
- `docs/config.mjs`, `docs/sidebar.mjs`, generated docs references
- docs package metadata and lockfiles
- Homebrew and AUR guidance
- GitHub workflow and release script references

Template-only wording will be removed. User-facing docs should describe `vx` as an actual CLI, not as a reusable template.
Repository hygiene should also match local conventions, including removing `npm`-specific lockfile leftovers if `bun` is the supported docs package manager.

## 5. Verification

Before deleting git history, verify:

- `just build`
- `just test`
- `go test ./...`
- targeted text search confirming no stale `go-cli-template` or `imdevan/go-cli-template` references remain, except where intentionally documented as provenance if kept at all

If generated docs depend on local tooling, also verify the docs generation path as far as the environment allows.

## 6. Git history reset

After code and docs are updated and verification passes:

1. Remove `.git`
2. Initialize a new repository
3. Leave the working tree ready for a fresh first commit under the new identity

This step is intentionally last so the migration can still benefit from diff-based verification during implementation.

## Risks

- Missed hardcoded references in docs or scripts can leave broken release paths.
- Generated docs may need regeneration after metadata changes.
- Package distribution helpers may contain assumptions about the old repository naming scheme.

## Mitigations

- Use repository-wide search before and after edits.
- Verify both build/test and textual identity cleanup.
- Keep history reset as the final step only after all checks pass.

## Testing Strategy

- Run build and test commands through the project `justfile`
- Run direct `go test ./...` as a second check
- Search for stale identifiers after edits
- Validate that the new binary path and command name are used consistently

## Success Criteria

The project is considered complete when:

- the codebase builds and tests under `github.com/vandordev/vx`
- the CLI entrypoint, docs, and scripts all use `vx`
- no operational references to the old template identity remain
- the repository no longer contains legacy git history
