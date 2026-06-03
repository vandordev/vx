# VX Prompted Template Inputs Design

## Goal

Define a small interactive input mode for `vx` template planning and generation.

The feature should let users fill missing `@input` values without typing long
flags, while keeping the default command behavior non-interactive and safe for
scripts.

The intended user-facing entrypoint is:

```bash
vx gen <target> -i
vx gen <target> --prompt
```

## Scope

This design covers:

- interactive prompting for missing `.vxt` template inputs
- `vx gen` and `vx generate`
- `vx view --plan`
- flag naming and conflict behavior
- prompt eligibility rules for TTY, JSON, and non-interactive execution
- precedence between `--values`, `--set`, and prompted values

The expected delivery target is the next `vx` minor or patch milestone after
`v0.3.1`.

## Non-Goals

This design does not cover:

- changing default `vx gen` behavior to interactive
- adding a new `vx geni` command
- prompting for injected `project.*` context
- full-screen TUI flows
- executing hooks
- changing `vxt` syntax or validation rules
- changing JSON output contracts except for preserving deterministic behavior

## UX Position

Default generation remains non-interactive.

Rationale:

- existing automation should not start waiting for input
- `vx gen` should remain predictable in scripts and CI
- users can opt into prompts with a short flag when working manually

The interactive mode should be explicit but cheap to type:

```text
-i, --prompt    prompt for missing template inputs
```

`--prompt` is preferred over `--interactive` because the behavior is specific:
it prompts for missing template inputs. It does not turn `vx` into a broader
interactive wizard.

## Command Behavior

### `vx gen <target>`

Without `-i` or `--prompt`, behavior stays as it is today.

If required input is missing:

- return the normal validation error
- do not prompt

### `vx gen <target> -i`

When input is missing and the command is attached to an interactive terminal:

1. load values from `--values`
2. apply `--set` overrides
3. inspect required template inputs
4. prompt only for missing user-declared `@input` fields
5. merge prompted values into the input map
6. run the normal generation planning flow
7. render the usual preview output

With `--apply`, the same prompting happens before planning and writing. Apply
still writes only after planning succeeds and after existing conflict checks
pass.

### `vx generate <target> -i`

`vx generate` is an alias for `vx gen`, so it should support exactly the same
prompt behavior.

### `vx view <target> --plan -i`

`vx view --plan -i` may prompt for missing inputs because planning needs
renderable values.

After prompting, it should render the normal planned files/directories output.

### `vx view <target> -i`

`vx view` without `--plan` should not prompt. It only inspects package/export
metadata and can already show required inputs without values.

If `-i` is supplied without `--plan`, either:

- ignore it with no prompt, or
- return a clear error that `--prompt` only applies with `--plan`

The preferred behavior is a clear error. It prevents users from thinking prompt
mode did something when no planning happened.

## Flag Conflicts

`--prompt` conflicts with deterministic or non-interactive modes.

These combinations should fail clearly before prompting:

```bash
vx gen <target> --json -i
vx gen <target> --non-interactive -i
vx view <target> --json -i
vx view <target> --non-interactive -i
```

Rationale:

- JSON output must be deterministic and machine-readable
- `--non-interactive` explicitly means fail instead of prompting
- conflicting flags should fail before any partial prompt state is created

## Input Precedence

Prompted values should have the lowest explicit precedence.

Final input precedence:

1. values loaded from `--values`
2. values supplied by `--set`
3. prompted values for keys that are still missing
4. injected `project.*` context merged by `vx`

Prompting should not overwrite a value already supplied by `--values` or
`--set`.

Injected `project.*` context remains runtime-owned and should not be prompted.
Templates should not declare `@input project`.

## Prompt Eligibility

Prompting should only happen when:

- `--prompt` or `-i` is present
- the command actually needs rendered input
- required user-declared input is missing
- stdin/stdout are attached to a terminal
- neither `--json` nor `--non-interactive` is present

If `--prompt` is requested outside a TTY, return a clear error such as:

```text
cannot prompt for input in a non-interactive terminal
```

## Prompt Experience

Prompts should be inline, not full-screen.

Use repository UI conventions:

- Bubble Tea/Bubbles/huh based components
- no full-screen TUI
- compact prompt labels
- type name and input type visible

Initial type handling can be conservative:

- `string`: text input
- `bool`: boolean confirm/select
- numeric types: text input parsed through the existing YAML scalar parser
- unsupported or complex object types: either prompt as YAML text or fail with
  a clear message in the first implementation

The prompt implementation should be easy to extend later for object fields and
select-like schemas if `vxt` grows richer input metadata.

## Architecture Boundary

Keep `cmd/vx` thin.

Recommended boundary:

- `cmd/vx`: flag wiring, conflict checks, TTY availability, and calling prompt
  orchestration
- `internal/input`: missing-input merge helpers and scalar parsing reuse
- `internal/ui`: inline prompt components
- `internal/gen` and `internal/view`: keep planning/generation services focused
  on validated input and runtime behavior

Do not embed prompt UI inside `internal/gen` or `internal/view`. Services should
remain testable without terminal UI.

## Error Handling

Errors should be early and explicit:

- missing input without `--prompt`: keep current missing-input behavior
- `--prompt` with `--json`: flag conflict
- `--prompt` with `--non-interactive`: flag conflict
- `--prompt` outside TTY: non-interactive terminal error
- invalid prompted scalar: show the input name and expected type
- interrupted prompt: return without planning or writing files

No files should be written if prompting fails or is cancelled.

## Testing Strategy

Unit and command-level tests should cover:

- `vx gen <target>` with missing input still fails without prompting
- `vx gen <target> -i` fills missing string input and previews planned output
- `vx gen <target> -i --apply` prompts before writing and writes expected files
- `vx generate <target> -i` matches `vx gen`
- `vx view <target> --plan -i` prompts and renders planned paths
- `vx view <target> -i` fails clearly without `--plan`
- `--values` and `--set` prevent prompts for already supplied keys
- prompted values do not overwrite `--set`
- `--json -i` fails before prompting
- `--non-interactive -i` fails before prompting
- non-TTY prompt attempts fail clearly
- cancellation does not write files

Tests should avoid relying on real terminal interactivity where possible. Use an
internal prompt interface with fake implementations for command tests, and keep
Bubble Tea/huh rendering tests focused on the UI package.

## Success Criteria

This design is successful when:

- users can type `vx gen <target> -i` to fill missing template inputs
- default `vx gen` remains non-interactive
- scripts and JSON consumers remain deterministic
- prompt behavior does not reach into generation service internals
- prompted values compose cleanly with `--values`, `--set`, and injected
  project context
