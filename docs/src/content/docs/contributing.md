---
title: Contributing
description: Contributing to vx
---


## Setup

```bash
gh repo clone vandordev/vx
cd vx
just build
just test
```

## Development

```bash
just build
just build-run
just watch
just test
just test-verbose
just docs-init
just docs-generate
```

## Pull Requests

1. Create a branch from `main`.
2. Keep changes scoped and update docs when behavior changes.
3. Run `just test` before opening the pull request.
4. Include a short explanation of what changed and why.
