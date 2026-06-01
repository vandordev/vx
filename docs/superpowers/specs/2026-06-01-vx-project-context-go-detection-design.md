# VX Project Context Go Detection Design

## Goal

Define a separate `vx` runtime design for injected project context, with an
initial focus on Go project detection and monorepo-safe module metadata.

This design is intended for the `vx v0.3.0` milestone rather than the existing
`v0.2.x` baseline.

This design should answer:

- what project context `vx` injects into templates
- where that context belongs in the product boundary
- how Go-specific context is detected
- how monorepo-safe module metadata should be exposed

## Scope

This design covers:

- injected project context owned by `vx`
- initial context fields for all detected projects
- Go-specific context fields
- detection rules for `go.mod`
- monorepo-safe naming for module metadata
- precedence rules for root detection vs language detection

The expected delivery target for this scope is `vx v0.3.0`.

## Non-Goals

This design does not cover:

- new `vxt` directives
- remote registry behavior
- non-Go language-specific context contracts
- full implementation planning for all future ecosystems

## Product Boundary

This context belongs in `vx`, not in `vxt`.

Rationale:

- `vxt` should remain a generic, typed generation engine
- project-aware and ecosystem-aware context is runtime knowledge
- adding language-specific directives to `vxt` would pollute the template
  language with execution-environment concerns

So the design position is:

- `vxt` stays generic
- `vx` detects project context
- `vx` injects detected context into the template runtime

## Initial Context Contract

The initial contract should be additive and sparse.

Base fields:

- `project.root`
- `project.language` when a language is detected

Go-specific fields:

- `project.go.module`
- `project.go.module_root`

Rules:

- `project.root` is always available when a `vx` command succeeds
- `project.language` is present only when a supported language is detected
- `project.go.*` fields are present only when Go is detected
- `vx` should not inject empty placeholders such as `project.go.module = ""`

If a context is not detected, it should be absent rather than blank.

## Why `project.go.module`, Not `project.module`

The Go module field should live under `project.go.module`, not `project.module`.

Rationale:

- avoids pretending every project has exactly one universal module concept
- avoids ambiguity in monorepos
- leaves room for future ecosystems without forcing a generic field too early
- keeps future growth predictable, for example:
  - `project.go.module`
  - `project.go.module_root`
  - `project.go.go_version`
  - later `project.node.*`, `project.python.*`, and similar fields

This is intentionally more explicit and more stable than a short generic field.

## Go Detection Rules

Go detection should be based on the nearest `go.mod` relative to the current
working directory.

It should not be based on:

- the target template path
- the first `go.mod` anywhere under the project root
- a global workspace-level module assumption

### Detection flow

1. detect `project.root` first from the nearest parent containing `vpkg/`
2. walk upward from `cwd` to find the nearest `go.mod`
3. ignore any `go.mod` outside `project.root`
4. if a valid in-root `go.mod` is found:
   - set `project.language = "go"`
   - parse `module` from `go.mod`
   - set `project.go.module`
   - set `project.go.module_root`

If no valid in-root `go.mod` is found:

- do not set `project.language = "go"`
- do not inject any `project.go.*` fields

## `project.go.module_root`

`project.go.module_root` should be the path relative to `project.root`, not an
absolute path.

Rationale:

- stable across machines
- safer for preview/apply output
- more useful inside templates
- better fit for monorepos

Example:

- `project.root = /repo`
- detected `go.mod = /repo/services/billing/go.mod`

Then:

- `project.go.module = github.com/acme/platform/services/billing`
- `project.go.module_root = services/billing`

## `project.language`

For the initial Go rollout, `project.language` should be a simple string:

- `go`

This should remain intentionally small for now.

Future multi-language expansion should add new fields rather than complicate the
initial contract prematurely.

## Availability Rules

This project context should be always-on if detected.

That means:

- `vx view` should receive it
- `vx gen` should receive it
- package/export type does not opt in explicitly
- direct `.vxt` sources also receive it when detected

The command does not need a special mode to enable project context.

## Monorepo Behavior

For monorepos, the nearest `go.mod` to `cwd` wins, as long as it remains inside
the detected `project.root`.

Example:

- `project.root = /repo`
- `cwd = /repo/services/billing/internal/app`
- nearest `go.mod = /repo/services/billing/go.mod`

Injected context:

- `project.root = /repo`
- `project.language = "go"`
- `project.go.module = "github.com/acme/platform/services/billing"`
- `project.go.module_root = "services/billing"`

This behavior is simple, predictable, and easy for users to control by where
they run the command.

## Success Criteria

This design is successful when:

- Go import/module metadata does not require new `vxt` directives
- `vx` owns project-aware detection cleanly
- the contract is safe for monorepos
- the initial field names are explicit enough to survive future language growth
