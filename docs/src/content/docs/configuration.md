---
title: Configuration
description: Configuration options for vx
---


Global configuration file location:

`$XDG_CONFIG_HOME/vx/config.toml`

Typical Linux path:

`~/.config/vx/config.toml`

Optional local override:

`./.vx/config.toml`

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `editor` | string | `nvim` | Editor opened by `vx config` |
| `list_spacing` | string | `space` | List density: `compact`, `tight`, or `space` |
| `headings` | string | `15` | Heading color |
| `primary` | string | `02` | Primary accent color |
| `secondary` | string | `06` | Secondary accent color |
| `text` | string | `07` | Body text color |
| `text_highlight` | string | `06` | Highlighted text color |
| `description_highlight` | string | `05` | Highlighted description color |
| `tags` | string | `13` | Tag color |
| `flags` | string | `12` | Flag and keybinding color |
| `muted` | string | `08` | Muted text color |
| `border` | string | `08` | Border color |

Colors accept named values, terminal palette indexes, or hex strings such as `"#ff8800"`.

## Example

```toml
# General
editor = "nvim"

# UI
list_spacing = "space"

# Colors
headings = "15"
primary = "02"
secondary = "06"
text = "07"
text_highlight = "06"
description_highlight = "05"
tags = "13"
flags = "12"
muted = "08"
border = "08"
```

## Commands

Create a config file:

```bash
vx config init
vx config init --force
vx config init --editor
```

Open the resolved config file in your configured editor:

```bash
vx config
```
