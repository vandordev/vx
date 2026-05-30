---
title: vx
description: A modern terminal-first CLI from Vandor Dev, built with Go, Cobra, and Bubble Tea.
---


`vx` is the main CLI project for `github.com/vandordev`.

The current codebase provides a terminal-first foundation built with Go, Cobra, and inline Bubble Tea components. Today it includes an interactive directory browser, configuration management, shell completion, generated documentation, and release automation for GitHub, Homebrew, and AUR.

## Features

- Inline Bubble Tea UI for browsing the current directory
- TOML configuration with XDG-aware global and local lookup
- Config bootstrap and editor integration
- Shell completion for bash, zsh, fish, and PowerShell
- Documentation generation with `gomarkdoc`, Astro Starlight, and Bun
- Packaging and release scripts for GitHub Releases, Homebrew, and AUR

## Requirements

- Go
- Just
- Bun

## Quick Start

```bash
gh repo clone vandordev/vx
cd vx
just build-run
```

## Commands

```bash
vx
vx config
vx config init
vx completion bash
```

`vx` currently opens an inline directory listing for the working directory. The supporting commands manage config files and shell completion.

## Configuration

Global configuration lives at `$XDG_CONFIG_HOME/vx/config.toml`, typically `~/.config/vx/config.toml`.

Local overrides can also be stored in:

```text
./.vx/config.toml
```

To initialize a config file:

```bash
vx config init
vx config init --force
vx config init --editor
```

To open the resolved config file in your editor:

```bash
vx config
```

See `example-config.toml` for the available keys.

## Development

```bash
just build
just build-run
just watch
just dev-build
just test
just test-verbose
just clean
```

## Documentation

The docs site lives in `docs/` and is built with Astro Starlight.

```bash
just docs-init
just docs-generate
just docs-dev
just docs-build
just docs-preview
just docs-clean
```

Generated docs pull from:

- root markdown files such as `README.md`, `INSTALL.md`, `CONFIG.md`, and `CONTRIBUTING.md`
- command metadata in `cmd/vx`
- API docs generated from packages under `internal/`

## Release Tooling

This repository keeps the packaging and release flows in place:

- `just github-release <version>`
- `just init-homebrew-tap`
- `just update-homebrew-formula <version>`
- `just init-aur-repo`
- `just update-aur-pkgbuild <version>`
- `just deploy-homebrew <version>`
- `just deploy-aur <version>`

## Project Layout

```text
.
├── cmd/vx
├── internal
├── docs
├── scripts
├── tests
└── justfile
```

- `cmd/vx` contains the Cobra entrypoint and subcommands.
- `internal/` contains domain models, config loading, UI, adapters, and utilities.
- `docs/` contains the Starlight documentation site.
- `scripts/` contains release and packaging automation.

## Installation

See `INSTALL.md` for installation options and release distribution notes.
