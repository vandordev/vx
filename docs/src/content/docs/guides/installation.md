---
title: Installation
description: How to install vx
---

## From Source

```bash
gh repo clone vandordev/vx
cd vx
just build
just install
```

## Using Go Install

```bash
go install github.com/vandordev/vx/cmd/vx@latest
```

## Homebrew

```bash
brew tap vandordev/homebrew-vx
brew install vx
```

## AUR

```bash
yay -S vx
```

## Verify Installation

```bash
vx --version
```

## Shell Completion

### Bash

```bash
vx completion bash > /etc/bash_completion.d/vx
```

### Zsh

```bash
vx completion zsh > "${fpath[1]}/_vx"
```

### Fish

```bash
vx completion fish > ~/.config/fish/completions/vx.fish
```

### PowerShell

```powershell
vx completion powershell | Out-String | Invoke-Expression
```
