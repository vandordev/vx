# Documentation Site

This directory contains the Astro Starlight site for `vx`.

## Local Development

```bash
cd docs
bun install
bun run dev
```

From the repository root you can also use:

```bash
just docs-init
just docs-generate
just docs-dev
```

## Content Sources

- `README.md`
- `INSTALL.md`
- `CONFIG.md`
- `CONTRIBUTING.md`
- `cmd/vx/*.go`
- `internal/**/*.go`
