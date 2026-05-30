---
title: Configuration
description: Configure vx
---

## Configuration File

`vx` uses TOML configuration files.

Global path:

```text
$XDG_CONFIG_HOME/vx/config.toml
```

Local override:

```text
./.vx/config.toml
```

## Initialize Configuration

```bash
vx config init
vx config init --force
vx config init --editor
```

## Open Configuration

```bash
vx config
```

## Configuration Reference

See the [example-config.toml](https://github.com/vandordev/vx/blob/main/example-config.toml) file for the available keys and defaults.
