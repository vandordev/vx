# Installation

## Quick Install

For macOS and Linux, install the latest release with:

```bash
curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

Install a pinned release:

```bash
VERSION=v0.3.0 curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

Override the install directory:

```bash
BIN_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/vandordev/vx/main/scripts/install.sh | sh
```

The installer defaults to `$HOME/.local/bin` and prints a PATH hint if that directory is not currently exported by your shell.

## From Source

```bash
gh repo clone vandordev/vx
cd vx
just build
just install
```

## With Go Install

`vx` is built from the executable package in `cmd/vx`, so the install path must target that package instead of the module root.

```bash
go install github.com/vandordev/vx/cmd/vx@latest
```

If you already cloned the repository locally:

```bash
go install ./cmd/vx
```

Do not use `go install github.com/vandordev/vx@latest`; the module root is not an executable `main` package.

Windows users should use:

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
