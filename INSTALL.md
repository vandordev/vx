# Installation

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
