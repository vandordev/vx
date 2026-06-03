# vx

A modern terminal-first CLI from Vandor Dev, built with Go, Cobra, and Bubble Tea.

The current public surface is project-local and preview-first. `vx` discovers packages from the nearest parent directory containing `vpkg/`, inspects packages and exports with `vx view`, and previews or applies `template` exports and direct `.vxt` files with `vx gen`.

## Features

- Project-root discovery through the nearest parent containing `vpkg/`
- Local package discovery from `vpkg/<namespace>/<package>/vpkg.yaml`
- `vx view` for package, export, and direct `.vxt` inspection
- `vx gen` and `vx generate` for preview-first template generation
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
just build
./bin/vx
```

## Commands

```bash
vx
vx view vandor/go-backend-core
vx view vandor/go-backend-core:default
vx view ./templates/usecase.vxt
vx gen vandor/go-backend-core --set name=create_booking
vx gen vandor/go-backend-core -i
vx gen ./templates/usecase.vxt --set context=booking --apply
vx gen vandor/go-backend-core -i --apply
vx generate vandor/go-backend-core:default --set name=create_booking
vx view vandor/go-backend-core:default --plan -i
vx config
vx config init
vx completion bash
```

`vx` without arguments prints an overview of the local runtime commands and examples.

`vx view` is non-destructive:

- package targets show identity, version, kind, and declared exports
- export targets show metadata and, for `template` exports, required inputs from the `.vxt` contract
- direct `.vxt` targets work when the file is inside the detected project root

`vx gen` is preview-first:

- only `template` exports and direct `.vxt` files are executable
- output writes target the detected project root, not the current working directory
- files are written only with `--apply`
- `-i, --prompt` prompts only for missing template `@input` values
- after prompted input, `vx gen -i` asks whether to `Preview` or `Apply`
- the action selector defaults to `Preview`
- `vx gen <target> -i --apply` skips the action selector and applies directly
- prompted values do not override values already supplied by `--values` or `--set`
- `--json` and `--non-interactive` cannot be combined with `--prompt`

## Project Layout

`vx` expects a project-local `vpkg/` tree such as:

```text
my-project/
├── vpkg/
│   └── vandor/
│       └── go-backend-core/
│           ├── vpkg.yaml
│           └── templates/
│               └── usecase.vxt
└── templates/
    └── standalone.vxt
```

Supported target forms:

- `namespace/package`
- `namespace/package:export`
- unique shorthand package such as `go-backend-core`
- unique shorthand export such as `usecase`
- direct path such as `./templates/usecase.vxt`

## Project Context

`vx` injects project context into `vxt` template input for `vx view --plan` and `vx gen`.

- `project.root` is always available for successful commands.
- `project.language` is currently only `go` and is present only when Go is detected.
- Go fields live under `project.go.*`.
- `project.go.module_root` is relative to `project.root`.
- The nearest in-root `go.mod` from the current working directory wins.
- Undetected context is omitted instead of injected as blank strings.

Example:

```vxt
@template service
@input name string
@file "{{ project.go.module_root }}/internal/{{ name | snake }}/service.go"
package {{ name | snake }}

type {{ name | pascal }}Service struct{}

// module: {{ project.go.module }}
@endfile
```

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

See `example-config.toml` for the available keys. Configuration does not control the root command behavior; runtime discovery comes from the local `vpkg/` project layout.

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

## Repository Layout

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
- `internal/` contains runtime services, config loading, package discovery, resolution, UI, adapters, and utilities.
- `docs/` contains the Starlight documentation site.
- `scripts/` contains release and packaging automation.

## Installation

For macOS and Linux, install the latest release with:

```bash
curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

Pin a release or change the install directory with:

```bash
VERSION=v0.4.0 curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
BIN_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

For Go users and Windows users, install the CLI from the executable package:

```bash
go install github.com/vandordev/vx/cmd/vx@latest
```

The repository root is a Go module, but the binary entrypoint lives in `cmd/vx`, so `go install github.com/vandordev/vx@latest` is not the correct install path.

See `INSTALL.md` for the full installation matrix and release distribution notes.
