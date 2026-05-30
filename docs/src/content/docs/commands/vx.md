---
title: vx
description: A modern terminal-first CLI from Vandor Dev, built with Go, Cobra, and Bubble Tea.
---

A modern terminal-first CLI from Vandor Dev, built with Go, Cobra, and Bubble Tea.

## Usage

```bash
vx [alias]
vx [command]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `-c, --config` | string | config file path |
| `-v, --version` | bool | print version information |

## Available Commands

- [`completion`](/commands/completion) - Generate shell completion scripts
- [`config`](/commands/config) - View or edit configuration
- [`config init`](/commands/config-init) - Generate a default config file

## Source

See [root.go](https://github.com/vandordev/vx/blob/main/cmd/vx/root.go) for implementation details.
