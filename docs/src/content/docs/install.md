---
title: Install
description: Installation instructions for vx
---


## From Source

```bash
gh repo clone vandordev/vx
cd vx
just build
just install
```

## With Go Install

```bash
go install github.com/vandordev/vx/cmd/vx@latest
```

## GitHub Releases

When release artifacts are published, download the archive for your platform from:

`https://github.com/vandordev/vx/releases`

Example for Linux AMD64:

```bash
curl -L https://github.com/vandordev/vx/releases/latest/download/vx-linux-amd64.tar.gz | tar -xz
sudo mv vx-linux-amd64 /usr/local/bin/vx
```

## Homebrew

Once the tap is published:

```bash
brew tap vandordev/homebrew-vx
brew install vx
```

## AUR

Once the package is published:

```bash
yay -S vx
```

## Verify

```bash
vx --version
```
