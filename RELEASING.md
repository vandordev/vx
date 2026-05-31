# Releasing

This file describes the release flow for `vx`.

## Prerequisites

- `gh auth status` must show a logged-in GitHub account with `repo` scope.
- `origin` should point to `https://github.com/vandordev/vx.git`.
- Work from a clean `main` branch.

## Minimum Release Flow

Use this flow when you want these install paths to pick up the new version:

- `go install github.com/vandordev/vx/cmd/vx@latest`
- `curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh`

### 1. Update the version source of truth

Edit `internal/package/package.toml`:

```toml
version = "0.3.0"
```

Then sync generated metadata:

```bash
just sync
```

### 2. Commit and push `main`

```bash
git add internal/package/package.toml
git commit -m "chore: release v0.3.0"
git push origin main
```

### 3. Create and push the tag

```bash
just tag 0.3.0
```

Important:

- Use `just tag 0.3.0`
- Do not use `just tag VERSION=0.3.0`

### 4. Publish the GitHub Release assets

```bash
just github-release 0.3.0
```

This builds and uploads the release archives used by `install.sh`.

## Optional Packaging Flow

Use this only when you also want to update package-manager distribution.

### 5. Update packaging metadata

```bash
just release 0.3.0
```

This updates:

- AUR `PKGBUILD`
- Homebrew formula source metadata

### 6. Publish Homebrew

```bash
just deploy-homebrew 0.3.0
```

### 7. Publish AUR

```bash
just deploy-aur 0.3.0
```

## Verification

Verify the new version from both public install paths.

### Go install

```bash
tmpbin=$(mktemp -d)
GOBIN="$tmpbin" go install github.com/vandordev/vx/cmd/vx@latest
"$tmpbin/vx" --version
```

### Installer script

```bash
tmpbin=$(mktemp -d)
curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | BIN_DIR="$tmpbin" sh
"$tmpbin/vx" --version
```

Both commands should print the release version you just published.

## Troubleshooting

### Wrong tag format

If you accidentally created a tag like `vVERSION=0.3.0`, delete it locally and remotely:

```bash
git tag -d vVERSION=0.3.0
git push origin :refs/tags/vVERSION=0.3.0
```

Then create the correct tag:

```bash
just tag 0.3.0
```

### Tag exists but release assets are missing

Pushing a `v*` tag triggers the GitHub Actions release workflow, but that workflow does not replace `just github-release`.

If the tag exists and the release page has no assets, run:

```bash
just github-release 0.3.0
```

### `go install @latest` and GitHub Release disagree on version

Check both:

- `internal/package/package.toml`
- the latest GitHub Release tag

They must refer to the same version before you announce the release.
