# VX Prompted Template Inputs Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit `-i, --prompt` mode so `vx gen`, `vx generate`, and `vx view --plan` can prompt for missing `.vxt` template inputs while default commands remain non-interactive.

**Architecture:** Keep `cmd/vx` thin by adding prompt orchestration helpers and fakeable prompt interfaces outside the generation/view services. Reuse existing `vxt` validation and planning by collecting missing input values before calling `internal/gen.Generate` or `internal/view.Inspect`.

**Tech Stack:** Go, Cobra flags, `github.com/charmbracelet/huh`/Bubble Tea for inline terminal prompts, existing `internal/input` YAML scalar parsing, existing `internal/ui` theme conventions, `just` for verification.

---

## Source Spec

Implement from:

- `docs/superpowers/specs/2026-06-03-vx-prompted-template-inputs-design.md`

Repository rules:

- Use `just`, not `make`.
- Keep `cmd/vx` thin.
- Put UI elements under `internal/ui`.
- Use inline Bubble Tea/Bubbles/huh prompts, not a full-screen TUI.
- Do not change default `vx gen` behavior to interactive.
- Do not add a `vx geni` command.

## File Structure

Create:

- `internal/input/missing.go`: helpers to inspect required inputs, detect which values are missing, merge prompted values without overwriting explicit values, and parse prompted scalars.
- `internal/input/missing_test.go`: unit tests for missing-input detection, dotted-key presence, merge precedence, and prompted scalar parsing.
- `internal/ui/input_prompt.go`: real inline prompt implementation for missing template inputs.
- `internal/ui/input_prompt_test.go`: focused tests for prompt request shaping and non-interactive parser helpers where possible.
- `cmd/vx/prompt.go`: command-level prompt interfaces, fakeable package variables, flag conflict checks, TTY checks, and orchestration shared by `gen` and `view`.
- `cmd/vx/prompt_test.go`: command-level tests for conflict checks and prompt orchestration without real terminal UI.

Modify:

- `cmd/vx/gen.go`: add `Prompt bool` option wired to `-i, --prompt`; call shared prompt orchestration before `gensvc.Generate`.
- `cmd/vx/view.go`: add `Prompt bool` option wired to `-i, --prompt`; reject `-i` without `--plan`; call shared prompt orchestration before `viewsvc.Inspect` only for planning.
- `cmd/vx/gen_test.go`: command integration tests for `gen -i`, `generate -i`, apply, conflicts, values/set precedence, and no default prompt.
- `cmd/vx/view_test.go`: command integration tests for `view --plan -i`, `view -i` error, and conflicts.
- `README.md`: document `-i, --prompt` in usage examples and behavior notes.
- `docs/src/content/docs/index.md`: generated from README by `just docs-generate`.
- `docs/src/content/docs/commands/gen.md`: generated command docs should reflect new flag after `just docs-generate`.
- `docs/src/content/docs/commands/view.md`: generated command docs should reflect new flag after `just docs-generate`.

Do not modify:

- `internal/gen/service.go` prompt behavior. It should still receive already-collected input.
- `internal/view/service.go` prompt behavior. It should still receive already-collected input.
- `vxt` dependency or template syntax.

## Chunk 1: Input Helper Foundation

### Task 1: Add Missing Input Detection

**Files:**

- Create: `internal/input/missing.go`
- Test: `internal/input/missing_test.go`

- [ ] **Step 1: Write failing tests for missing input detection**

Add tests:

```go
func TestMissingRequiredInputs(t *testing.T) {
	fields := []input.RequiredField{
		{Name: "name", TypeName: "string"},
		{Name: "feature.enabled", TypeName: "bool"},
		{Name: "project", TypeName: "object"},
	}
	values := map[string]any{
		"name": "booking",
		"feature": map[string]any{"enabled": true},
	}

	missing := input.MissingRequiredInputs(fields, values)
	if len(missing) != 0 {
		t.Fatalf("missing = %#v", missing)
	}
}

func TestMissingRequiredInputsSkipsProjectInput(t *testing.T) {
	fields := []input.RequiredField{{Name: "project", TypeName: "object"}}

	missing := input.MissingRequiredInputs(fields, map[string]any{})
	if len(missing) != 0 {
		t.Fatalf("missing = %#v", missing)
	}
}
```

Also cover:

- top-level value missing
- dotted value missing when parent map is absent
- dotted value missing when leaf is absent
- zero values such as `false`, `0`, and `""` count as present

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/input -run 'TestMissingRequiredInputs' -v
```

Expected: FAIL because `RequiredField` or `MissingRequiredInputs` does not exist.

- [ ] **Step 3: Implement required field helper**

In `internal/input/missing.go`, add:

```go
package input

import "strings"

type RequiredField struct {
	Name     string
	TypeName string
}

func MissingRequiredInputs(fields []RequiredField, values map[string]any) []RequiredField {
	var missing []RequiredField
	for _, field := range fields {
		if field.Name == "project" || strings.HasPrefix(field.Name, "project.") {
			continue
		}
		if !hasValue(values, field.Name) {
			missing = append(missing, field)
		}
	}
	return missing
}

func hasValue(values map[string]any, name string) bool {
	if values == nil {
		return false
	}
	parts := strings.Split(name, ".")
	var current any = values
	for _, part := range parts {
		object, ok := current.(map[string]any)
		if !ok {
			return false
		}
		next, ok := object[part]
		if !ok {
			return false
		}
		current = next
	}
	return true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/input -run 'TestMissingRequiredInputs' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/input/missing.go internal/input/missing_test.go
git commit -m "feat: detect missing template inputs"
```

### Task 2: Add Prompted Value Parsing And Merge Helpers

**Files:**

- Modify: `internal/input/missing.go`
- Test: `internal/input/missing_test.go`

- [ ] **Step 1: Write failing tests for prompted merge precedence**

Add tests:

```go
func TestMergePromptedValuesDoesNotOverwriteExistingValues(t *testing.T) {
	values := map[string]any{"name": "from-set"}
	prompted := map[string]any{"name": "from-prompt", "count": 2}

	merged := input.MergePromptedValues(values, prompted)

	if merged["name"] != "from-set" {
		t.Fatalf("name = %#v", merged["name"])
	}
	if merged["count"] != 2 {
		t.Fatalf("count = %#v", merged["count"])
	}
}

func TestParsePromptValueUsesInputType(t *testing.T) {
	value, err := input.ParsePromptValue("enabled", "bool", "true")
	if err != nil {
		t.Fatalf("ParsePromptValue returned error: %v", err)
	}
	if value != true {
		t.Fatalf("value = %#v", value)
	}
}
```

Also cover:

- dotted prompted key merges into nested map
- invalid bool returns an error containing the input name
- string values stay strings even when they look numeric
- numeric input types use YAML scalar parsing

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/input -run 'TestMergePromptedValues|TestParsePromptValue' -v
```

Expected: FAIL because helpers do not exist.

- [ ] **Step 3: Implement minimal helpers**

In `internal/input/missing.go`, add:

```go
func MergePromptedValues(values map[string]any, prompted map[string]any) map[string]any {
	merged := cloneMap(values)
	for key, value := range prompted {
		if !hasValue(merged, key) {
			setDottedValue(merged, key, value)
		}
	}
	return merged
}

func ParsePromptValue(name, typeName, raw string) (any, error) {
	switch typeName {
	case "string":
		return raw, nil
	default:
		value, err := parseScalar(raw)
		if err != nil {
			return nil, fmt.Errorf("parse prompted value for %q: %w", name, err)
		}
		return value, nil
	}
}
```

Add small private helpers:

- `cloneMap`
- `setDottedValue`

Reuse existing unexported `parseScalar` from `internal/input/values.go`.

- [ ] **Step 4: Run package tests**

Run:

```bash
go test ./internal/input -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/input/missing.go internal/input/missing_test.go
git commit -m "feat: merge prompted template input values"
```

## Chunk 2: Prompt UI Boundary

### Task 3: Add Prompt Interface And Real UI Implementation

**Files:**

- Create: `internal/ui/input_prompt.go`
- Test: `internal/ui/input_prompt_test.go`

- [ ] **Step 1: Write focused tests for prompt request shaping**

Add tests for pure helpers only, not a real terminal session:

```go
func TestInputPromptLabelsIncludeNameAndType(t *testing.T) {
	fields := []ui.PromptField{{Name: "name", TypeName: "string"}}

	labels := ui.InputPromptLabels(fields)

	if labels[0] != "name (string)" {
		t.Fatalf("label = %q", labels[0])
	}
}
```

If the implementation exposes different pure helpers, keep the test equivalent:
the labels used by the prompt must include both name and type.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/ui -run 'TestInputPrompt' -v
```

Expected: FAIL because prompt types/helpers do not exist.

- [ ] **Step 3: Implement inline prompt UI**

In `internal/ui/input_prompt.go`, add:

```go
package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

type PromptField struct {
	Name     string
	TypeName string
}

type PromptValue struct {
	Name  string
	Value string
}

func PromptInputs(fields []PromptField, theme Theme) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	formFields := make([]huh.Field, 0, len(fields))
	bindings := make([]struct {
		name  string
		value *string
	}, 0, len(fields))
	for _, field := range fields {
		value := ""
		bindings = append(bindings, struct {
			name  string
			value *string
		}{name: field.Name, value: &value})
		formFields = append(formFields, huh.NewInput().
			Title(promptLabel(field)).
			Value(&value))
	}
	if err := huh.NewForm(huh.NewGroup(formFields...)).
		WithTheme(inputPromptTheme(theme)).
		Run(); err != nil {
		return nil, err
	}
	for _, binding := range bindings {
		values[binding.name] = *binding.value
	}
	return values, nil
}
```

Implementation notes:

- The exact helper names may vary.
- If `huh.Input.Value` requires `*string`, keep a local slice of `{name, *string}`
  bindings and copy them into the result map after `Run`.
- Use `huh.ThemeBase()` with minimal `lipgloss` color customization matching
  `internal/ui/confirmation.go`.
- Keep the form inline. Do not add fullscreen Bubble Tea program options.

- [ ] **Step 4: Run UI tests**

Run:

```bash
go test ./internal/ui -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/input_prompt.go internal/ui/input_prompt_test.go
git commit -m "feat: add inline template input prompts"
```

## Chunk 3: Command Prompt Orchestration

### Task 4: Add Shared Command Prompt Orchestration

**Files:**

- Create: `cmd/vx/prompt.go`
- Test: `cmd/vx/prompt_test.go`

- [ ] **Step 1: Write failing tests for flag conflict validation**

Add tests:

```go
func TestValidatePromptOptionsRejectsJSON(t *testing.T) {
	err := validatePromptOptions(promptOptions{Prompt: true, AsJSON: true})
	if err == nil || !strings.Contains(err.Error(), "--prompt cannot be used with --json") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidatePromptOptionsRejectsNonInteractive(t *testing.T) {
	err := validatePromptOptions(promptOptions{Prompt: true, NonInteractive: true})
	if err == nil || !strings.Contains(err.Error(), "--prompt cannot be used with --non-interactive") {
		t.Fatalf("error = %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./cmd/vx -run 'TestValidatePromptOptions' -v
```

Expected: FAIL because helpers do not exist.

- [ ] **Step 3: Implement conflict helper and fakeable dependencies**

In `cmd/vx/prompt.go`, add:

```go
type promptOptions struct {
	Prompt         bool
	AsJSON         bool
	NonInteractive bool
	RequiresPlan   bool
	CommandName     string
}

func validatePromptOptions(opts promptOptions) error {
	if !opts.Prompt {
		return nil
	}
	if opts.AsJSON {
		return fmt.Errorf("--prompt cannot be used with --json")
	}
	if opts.NonInteractive {
		return fmt.Errorf("--prompt cannot be used with --non-interactive")
	}
	if !opts.RequiresPlan {
		return fmt.Errorf("--prompt requires template planning")
	}
	return nil
}
```

Also define package variables for tests:

```go
var stdinIsTerminal = func() bool { return tty.IsTerminal(os.Stdin.Fd()) }
var stdoutIsTerminal = func() bool { return tty.IsTerminal(os.Stdout.Fd()) }
var promptInputs = realPromptInputs
```

Use `internal/adapters/tty` for terminal checks.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./cmd/vx -run 'TestValidatePromptOptions' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/vx/prompt.go cmd/vx/prompt_test.go
git commit -m "feat: validate vx prompt mode flags"
```

### Task 5: Resolve Missing Inputs Before Generation/View

**Files:**

- Modify: `cmd/vx/prompt.go`
- Test: `cmd/vx/prompt_test.go`

- [ ] **Step 1: Write failing tests for fake prompt orchestration**

Add tests that do not use real terminal UI:

```go
func TestResolvePromptedInputPromptsOnlyMissingFields(t *testing.T) {
	restore := stubPrompt(t, map[string]string{"name": "booking"})
	defer restore()
	restoreTTY := stubTTY(t, true, true)
	defer restoreTTY()

	values, err := resolvePromptedInput(map[string]any{"count": 2}, []input.RequiredField{
		{Name: "name", TypeName: "string"},
		{Name: "count", TypeName: "number"},
	}, true)
	if err != nil {
		t.Fatalf("resolvePromptedInput returned error: %v", err)
	}
	if values["name"] != "booking" || values["count"] != 2 {
		t.Fatalf("values = %#v", values)
	}
}
```

Also cover:

- no prompt call when no required inputs are missing
- non-TTY returns `cannot prompt for input in a non-interactive terminal`
- prompted scalar parse errors include input name
- prompt never requests `project` or `project.go.module`

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./cmd/vx -run 'TestResolvePromptedInput' -v
```

Expected: FAIL because orchestration helper does not exist.

- [ ] **Step 3: Implement prompt resolution helper**

In `cmd/vx/prompt.go`, add a helper with this shape:

```go
func resolvePromptedInput(values map[string]any, fields []input.RequiredField, prompt bool) (map[string]any, error) {
	if !prompt {
		return values, nil
	}
	missing := input.MissingRequiredInputs(fields, values)
	if len(missing) == 0 {
		return values, nil
	}
	if !stdinIsTerminal() || !stdoutIsTerminal() {
		return nil, fmt.Errorf("cannot prompt for input in a non-interactive terminal")
	}

	raw, err := promptInputs(toPromptFields(missing))
	if err != nil {
		return nil, err
	}

	parsed := map[string]any{}
	for _, field := range missing {
		value, err := input.ParsePromptValue(field.Name, field.TypeName, raw[field.Name])
		if err != nil {
			return nil, err
		}
		parsed[field.Name] = value
	}
	return input.MergePromptedValues(values, parsed), nil
}
```

The real `promptInputs` adapter should call `ui.PromptInputs(fields, ui.ThemeFromConfig(domain.Config{}))`.

- [ ] **Step 4: Run prompt tests**

Run:

```bash
go test ./cmd/vx -run 'TestResolvePromptedInput|TestValidatePromptOptions' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/vx/prompt.go cmd/vx/prompt_test.go
git commit -m "feat: resolve prompted vx command inputs"
```

## Chunk 4: Wire `vx gen` And `vx generate`

### Task 6: Add `-i, --prompt` To `vx gen`

**Files:**

- Modify: `cmd/vx/gen.go`
- Test: `cmd/vx/gen_test.go`

- [ ] **Step 1: Write failing command tests for `gen -i`**

Add tests in `cmd/vx/gen_test.go` using package-level stubs from `prompt.go`:

```go
func TestGenCommandPromptMode(t *testing.T) {
	// fixture with @input name string
	// stub TTY true
	// stub prompt returns map[string]string{"name": "booking"}
	output, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "-i")
	if err != nil {
		t.Fatalf("gen -i returned error: %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "internal/booking.txt") {
		t.Fatalf("output = %s", output)
	}
}
```

Also add tests:

- `generate <target> -i` matches `gen <target> -i`
- `gen <target> -i --apply` writes generated file
- `gen <target>` without `-i` and missing input still fails
- `gen <target> --json -i` fails with conflict before prompt
- `gen <target> --non-interactive -i` fails with conflict before prompt
- `--set name=from-set -i` does not call prompt for `name`
- `--values values.yaml -i` does not call prompt for supplied values

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/vx -run 'TestGenCommandPromptMode|TestGenCommandOutput' -v
```

Expected: FAIL because `-i`/`--prompt` is not wired.

- [ ] **Step 3: Wire prompt flag and required input extraction**

In `cmd/vx/gen.go`:

- Add `prompt bool` to `genOptions`.
- Register:

```go
cmd.Flags().BoolVarP(&opts.prompt, "prompt", "i", false, "prompt for missing template inputs")
```

- After resolving target and loading values, but before detecting project context or calling `gensvc.Generate`, validate prompt options:

```go
if err := validatePromptOptions(promptOptions{
	Prompt: opts.prompt,
	AsJSON: opts.asJSON,
	NonInteractive: opts.nonInteractive,
	RequiresPlan: true,
	CommandName: "gen",
}); err != nil {
	return err
}
```

- Compile/inspect required inputs for the resolved target without planning output. Prefer adding a small command helper in `prompt.go` that uses `viewsvc.Inspect` without `Plan` to get `RequiredInputs`, then converts to `input.RequiredField`.
- Call `resolvePromptedInput(values, fields, opts.prompt)`.
- Pass the merged values to `gensvc.Generate`.

Keep `internal/gen` unchanged.

- [ ] **Step 4: Run gen command tests**

Run:

```bash
go test ./cmd/vx -run 'TestGenCommandPromptMode|TestGenCommandOutput|TestGenerateAlias' -v
```

Expected: PASS.

- [ ] **Step 5: Run focused service tests to catch regressions**

Run:

```bash
go test ./internal/gen ./cmd/vx -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/vx/gen.go cmd/vx/gen_test.go cmd/vx/prompt.go cmd/vx/prompt_test.go
git commit -m "feat: prompt for missing vx gen inputs"
```

## Chunk 5: Wire `vx view --plan`

### Task 7: Add `-i, --prompt` To `vx view --plan`

**Files:**

- Modify: `cmd/vx/view.go`
- Test: `cmd/vx/view_test.go`

- [ ] **Step 1: Write failing command tests for `view --plan -i`**

Add tests:

```go
func TestViewCommandPromptMode(t *testing.T) {
	// fixture with @input name string
	// stub TTY true
	// stub prompt returns map[string]string{"name": "booking"}
	output, err := testutil.RunCLI(t, newRootCmd(), "view", "vandor/go-backend-core:default", "--plan", "-i")
	if err != nil {
		t.Fatalf("view --plan -i returned error: %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "internal/booking.txt") {
		t.Fatalf("output = %s", output)
	}
}
```

Also add tests:

- `view <target> -i` fails with message that prompt requires planning
- `view <target> --plan --json -i` fails before prompt
- `view <target> --non-interactive -i` fails before prompt
- `view <target> --plan` without `-i` still fails on missing input
- supplied `--set` values do not prompt

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/vx -run 'TestViewCommandPromptMode|TestViewCommandOutput' -v
```

Expected: FAIL because `-i`/`--prompt` is not wired for view.

- [ ] **Step 3: Wire prompt flag and planning-only validation**

In `cmd/vx/view.go`:

- Add `prompt bool` to `viewOptions`.
- Register:

```go
cmd.Flags().BoolVarP(&opts.prompt, "prompt", "i", false, "prompt for missing template inputs when planning")
```

- Validate:

```go
if err := validatePromptOptions(promptOptions{
	Prompt: opts.prompt,
	AsJSON: opts.asJSON,
	NonInteractive: opts.nonInteractive,
	RequiresPlan: opts.plan,
	CommandName: "view",
}); err != nil {
	return err
}
```

- Only when `opts.plan` is true, resolve missing inputs before calling `viewsvc.Inspect`.
- Keep `view` without `--plan` non-prompting and erroring if `-i` is supplied.

- [ ] **Step 4: Run view command tests**

Run:

```bash
go test ./cmd/vx -run 'TestViewCommandPromptMode|TestViewCommandOutput|TestViewCommandInjectsProjectContextForPlan' -v
```

Expected: PASS.

- [ ] **Step 5: Run focused view tests**

Run:

```bash
go test ./internal/view ./cmd/vx -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/vx/view.go cmd/vx/view_test.go cmd/vx/prompt.go cmd/vx/prompt_test.go
git commit -m "feat: prompt for missing vx view plan inputs"
```

## Chunk 6: Documentation And Full Verification

### Task 8: Document Prompt Mode

**Files:**

- Modify: `README.md`
- Generated: `docs/src/content/docs/index.md`
- Generated: `docs/src/content/docs/commands/gen.md`
- Generated: `docs/src/content/docs/commands/view.md`

- [ ] **Step 1: Update README examples and behavior notes**

Add examples near command workflow:

```bash
vx gen vandor/go-backend-core -i
vx gen vandor/go-backend-core -i --apply
vx view vandor/go-backend-core:default --plan -i
```

Add behavior notes:

- `-i, --prompt` prompts for missing template inputs.
- Default `vx gen` remains non-interactive.
- `--json` and `--non-interactive` cannot be combined with `--prompt`.
- Prompted values fill only missing `@input` values and do not override
  `--values` or `--set`.

- [ ] **Step 2: Regenerate docs**

Run:

```bash
just docs-generate
```

Expected: PASS and generated docs update command flag pages.

- [ ] **Step 3: Inspect docs diff**

Run:

```bash
git diff -- README.md docs/src/content/docs/index.md docs/src/content/docs/commands/gen.md docs/src/content/docs/commands/view.md
```

Expected: Diff documents `-i, --prompt` only; no unrelated generated churn.

- [ ] **Step 4: Commit docs**

```bash
git add README.md docs/src/content/docs/index.md docs/src/content/docs/commands/gen.md docs/src/content/docs/commands/view.md
git commit -m "docs: document vx prompt mode"
```

### Task 9: Full Verification

**Files:**

- Verify all changed files.

- [ ] **Step 1: Run full Go tests**

Run:

```bash
just test
```

Expected: PASS.

- [ ] **Step 2: Run build**

Run:

```bash
just build
```

Expected: PASS and `bin/vx` is built.

- [ ] **Step 3: Run docs generation**

Run:

```bash
just docs-generate
```

Expected: PASS and no additional diff after generated docs have been committed.

- [ ] **Step 4: Run docs build as known-risk verification**

Run:

```bash
just docs-build
```

Expected: This may still fail at `@astrojs/sitemap` with
`Cannot read properties of undefined (reading 'reduce')`, which is a known
non-blocking docs toolchain issue from before this feature. If it fails, capture
the exact failure in the final handoff. If it passes, report that the known issue
is no longer reproducing.

- [ ] **Step 5: Run whitespace check**

Run:

```bash
git diff --check
```

Expected: PASS with no output.

- [ ] **Step 6: Manual smoke test with built binary**

Create a temporary fixture or use an existing test fixture project, then run:

```bash
./bin/vx gen vandor/go-backend-core -i
```

Expected: In an interactive terminal, the command prompts for missing inputs,
then renders the usual preview. Do not commit generated smoke-test files.

- [ ] **Step 7: Commit any final fixes**

If verification requires fixes:

```bash
git add <changed-files>
git commit -m "fix: harden vx prompt mode"
```

- [ ] **Step 8: Final status check**

Run:

```bash
git status --short
```

Expected: clean worktree, or only intentionally uncommitted local smoke-test
artifacts that have been removed before handoff.

## Integration Notes

- This feature builds on `vxt` required input metadata exposed through compiled
  documents and current `viewsvc.Inspect` required input summaries.
- Do not change the current `vxt` case filter behavior.
- Do not change injected project context behavior from `v0.3.0`.
- Do not treat `project.*` fields as promptable user input.
- `--non-interactive` already exists in `gen` and `view`; this plan makes it
  meaningful as a conflict with explicit prompt mode, but default behavior
  remains non-interactive.

## Execution Handoff

When executing this plan:

1. Use TDD task-by-task.
2. Commit after each task as listed.
3. Keep services free of terminal UI dependencies.
4. Prefer focused command helpers over large command functions.
5. Preserve existing behavior unless a task explicitly changes it.
