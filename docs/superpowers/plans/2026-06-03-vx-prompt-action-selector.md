# VX Prompt Action Selector Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let `vx gen <target> -i` ask whether to preview or apply after it has prompted for missing template inputs, so manual generation can finish without typing `--apply` up front.

**Architecture:** Treat this as a narrow amendment to explicit prompt mode from `docs/superpowers/specs/2026-06-03-vx-prompted-template-inputs-design.md`. Keep generation services unchanged: command orchestration decides whether an interactive prompt session happened, asks for the final action, then calls the existing preview or apply path.

**Tech Stack:** Go, Cobra command wiring, `github.com/charmbracelet/huh` inline select UI through `internal/ui`, existing `cmd/vx` prompt orchestration, `just` for verification.

---

## Scope

Implement only for:

- `vx gen <target> -i`
- `vx generate <target> -i`

Do not implement for:

- `vx view --plan -i`, because view planning never writes files.
- default non-interactive `vx gen`.
- `--json` or `--non-interactive` modes, which already conflict with `--prompt`.

Behavior to preserve:

- `vx gen <target> -i --apply` remains explicit apply and does not show an action selector.
- `vx gen <target> -i` shows the action selector only when a missing-input prompt actually ran.
- If there are no missing inputs, `vx gen <target> -i` stays a normal preview command.
- The action selector default is `Preview`.
- Selecting `Apply` should behave exactly like passing `--apply`, including existing conflict checks and existing-file protection.
- Prompt cancellation or action cancellation must not plan or write files.

Naming:

- Use action labels `Preview` and `Apply`.
- Avoid `Dry`, `Dry run`, or `Applied`; `Preview` matches existing vx wording.

## File Structure

Create:

- `internal/ui/action_select.go`: reusable inline action selector for prompt-backed command decisions.
- `internal/ui/action_select_test.go`: tests for action option definitions and default action behavior without launching a real TUI.

Modify:

- `cmd/vx/prompt.go`: extend prompt orchestration result so command code can know whether missing-input prompts were shown; add a fakeable action selector adapter for tests.
- `cmd/vx/prompt_test.go`: unit tests for action selector orchestration, cancellation, and skip conditions.
- `cmd/vx/gen.go`: after input prompting and before `gensvc.Generate`, resolve final preview/apply intent from `--apply` plus the optional action selector.
- `cmd/vx/gen_test.go`: command integration tests for `gen -i` action choices, alias behavior, and no-selector cases.
- `README.md`: document that `vx gen <target> -i` can choose Preview or Apply after prompted input.
- `docs/src/content/docs/index.md`: generated from README by `just docs-generate`.
- `docs/src/content/docs/commands/gen.md`: generated command docs should remain correct after `just docs-generate`.

Do not modify:

- `internal/gen/service.go`: it should still receive final generation options after command orchestration.
- `internal/view/service.go`: no action selector applies to view.
- `internal/input/missing.go`: missing-input detection and merge precedence should remain unchanged unless tests expose a real gap.

## Chunk 1: UI Action Selector

### Task 1: Add Action Selector UI Type

**Files:**

- Create: `internal/ui/action_select.go`
- Test: `internal/ui/action_select_test.go`

- [ ] **Step 1: Write failing tests for action defaults**

Add tests that do not start a real terminal UI:

```go
func TestGenerationActionOptions(t *testing.T) {
	options := ui.GenerationActionOptions()

	if len(options) != 2 {
		t.Fatalf("len(options) = %d, want 2", len(options))
	}
	if options[0].Value != ui.GenerationActionPreview {
		t.Fatalf("first option = %q, want preview", options[0].Value)
	}
	if options[1].Value != ui.GenerationActionApply {
		t.Fatalf("second option = %q, want apply", options[1].Value)
	}
}

func TestDefaultGenerationAction(t *testing.T) {
	if got := ui.DefaultGenerationAction(); got != ui.GenerationActionPreview {
		t.Fatalf("DefaultGenerationAction() = %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/ui -run 'TestGenerationAction' -v
```

Expected: FAIL because the action type and helpers do not exist.

- [ ] **Step 3: Implement minimal selector helpers**

In `internal/ui/action_select.go`, add:

```go
package ui

import "github.com/charmbracelet/huh"

type GenerationAction string

const (
	GenerationActionPreview GenerationAction = "preview"
	GenerationActionApply   GenerationAction = "apply"
)

type GenerationActionOption struct {
	Label string
	Value GenerationAction
}

func GenerationActionOptions() []GenerationActionOption {
	return []GenerationActionOption{
		{Label: "Preview", Value: GenerationActionPreview},
		{Label: "Apply", Value: GenerationActionApply},
	}
}

func DefaultGenerationAction() GenerationAction {
	return GenerationActionPreview
}

func PromptGenerationAction() (GenerationAction, error) {
	action := DefaultGenerationAction()
	return action, huh.NewSelect[GenerationAction]().
		Title("What should vx do next?").
		Options(
			huh.NewOption("Preview", GenerationActionPreview),
			huh.NewOption("Apply", GenerationActionApply),
		).
		Value(&action).
		Run()
}
```

Keep the prompt inline. Do not introduce a full-screen TUI.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/ui -run 'TestGenerationAction' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/action_select.go internal/ui/action_select_test.go
git commit -m "feat: add generation action selector UI"
```

## Chunk 2: Prompt Orchestration Contract

### Task 2: Track Whether Prompting Happened

**Files:**

- Modify: `cmd/vx/prompt.go`
- Test: `cmd/vx/prompt_test.go`

- [ ] **Step 1: Write failing tests for prompt result metadata**

Add or extend tests around the existing fake prompt adapter:

```go
func TestCompletePromptedInputsReportsPromptedWhenValuesWereMissing(t *testing.T) {
	// Arrange a required input missing from values and a fake prompt that returns it.
	// Assert the returned result has Prompted == true.
}

func TestCompletePromptedInputsReportsNotPromptedWhenNothingMissing(t *testing.T) {
	// Arrange all required inputs present.
	// Assert the returned result has Prompted == false and the fake prompt was not called.
}
```

Use existing command-test helpers in `cmd/vx/prompt_test.go`; do not invoke real TTY UI.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./cmd/vx -run 'TestCompletePromptedInputsReports' -v
```

Expected: FAIL because prompt completion currently returns only values or does not expose whether a prompt ran.

- [ ] **Step 3: Implement prompt result metadata**

In `cmd/vx/prompt.go`, add a small result type if one does not already exist:

```go
type promptedInputResult struct {
	Values   map[string]any
	Prompted bool
}
```

Update the shared input prompt orchestration to:

- return `Prompted: false` when `--prompt` is disabled
- return `Prompted: false` when no user-declared `@input` is missing
- return `Prompted: true` only after the missing-input UI successfully collects values
- preserve current error behavior for conflicts, non-TTY, invalid values, and cancellation

Keep callers simple: they should receive the merged values and the prompted flag.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./cmd/vx -run 'TestCompletePromptedInputsReports' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/vx/prompt.go cmd/vx/prompt_test.go
git commit -m "feat: report prompted input sessions"
```

### Task 3: Add Fakeable Action Selection Boundary

**Files:**

- Modify: `cmd/vx/prompt.go`
- Test: `cmd/vx/prompt_test.go`

- [ ] **Step 1: Write failing tests for action selection rules**

Add tests:

```go
func TestResolvePromptedGenerationApplyUsesExplicitApply(t *testing.T) {
	apply, err := resolvePromptedGenerationApply(promptedGenerationApplyOptions{
		ExplicitApply: true,
		Prompted:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !apply {
		t.Fatalf("apply = false, want true")
	}
}

func TestResolvePromptedGenerationApplySkipsSelectorWhenNotPrompted(t *testing.T) {
	apply, err := resolvePromptedGenerationApply(promptedGenerationApplyOptions{
		ExplicitApply: false,
		Prompted:      false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apply {
		t.Fatalf("apply = true, want false")
	}
}
```

Also cover:

- `Prompted: true`, selector returns preview -> apply false
- `Prompted: true`, selector returns apply -> apply true
- selector error returns error and does not choose apply
- selector is not called when `ExplicitApply` is true

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./cmd/vx -run 'TestResolvePromptedGenerationApply' -v
```

Expected: FAIL because the resolver does not exist.

- [ ] **Step 3: Implement resolver and adapter**

In `cmd/vx/prompt.go`, add a package-level adapter similar to the existing fakeable prompt input adapter:

```go
var promptGenerationAction = ui.PromptGenerationAction

type promptedGenerationApplyOptions struct {
	ExplicitApply bool
	Prompted      bool
}

func resolvePromptedGenerationApply(opts promptedGenerationApplyOptions) (bool, error) {
	if opts.ExplicitApply {
		return true, nil
	}
	if !opts.Prompted {
		return false, nil
	}
	action, err := promptGenerationAction()
	if err != nil {
		return false, err
	}
	return action == ui.GenerationActionApply, nil
}
```

Reset the package-level adapter in tests with `t.Cleanup`.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./cmd/vx -run 'TestResolvePromptedGenerationApply' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/vx/prompt.go cmd/vx/prompt_test.go
git commit -m "feat: resolve prompted generation action"
```

## Chunk 3: Wire `vx gen`

### Task 4: Apply Selector Decision In `vx gen`

**Files:**

- Modify: `cmd/vx/gen.go`
- Test: `cmd/vx/gen_test.go`

- [ ] **Step 1: Write failing command tests**

Add tests:

```go
func TestGenPromptedActionApplyWritesFiles(t *testing.T) {
	// Arrange a template requiring @input name.
	// Run: vx gen <target> -i
	// Fake missing-input prompt returns name=booking.
	// Fake action selector returns Apply.
	// Assert planned file exists on disk.
}

func TestGenPromptedActionPreviewDoesNotWriteFiles(t *testing.T) {
	// Same setup, but fake action selector returns Preview.
	// Assert preview output is shown and file does not exist.
}
```

Also cover:

- `vx gen <target> -i --apply` writes without calling the action selector
- `vx gen <target> -i --set name=booking` does not call the action selector because no missing-input prompt ran
- existing file conflict still fails when action selector chooses apply

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./cmd/vx -run 'TestGenPromptedAction' -v
```

Expected: FAIL because `vx gen -i` always previews unless `--apply` is supplied.

- [ ] **Step 3: Wire selector into generation options**

In `cmd/vx/gen.go`:

1. Capture the prompted-input result from shared prompt orchestration.
2. Call `resolvePromptedGenerationApply` with:

```go
promptedGenerationApplyOptions{
	ExplicitApply: opts.Apply,
	Prompted:      promptedResult.Prompted,
}
```

3. Pass the resolved apply value to `gensvc.Generate`.
4. Return selector errors before calling `gensvc.Generate`.

Do not move this logic into `internal/gen`.

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
go test ./cmd/vx -run 'TestGenPromptedAction|TestGenPrompt' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/vx/gen.go cmd/vx/gen_test.go
git commit -m "feat: choose prompted generation action"
```

### Task 5: Verify `vx generate` Alias Behavior

**Files:**

- Modify: `cmd/vx/gen_test.go`

- [ ] **Step 1: Write failing alias test if coverage is missing**

Add:

```go
func TestGeneratePromptedActionApplyWritesFiles(t *testing.T) {
	// Same behavior as vx gen, but invoke the generate alias.
}
```

- [ ] **Step 2: Run test**

Run:

```bash
go test ./cmd/vx -run 'TestGeneratePromptedAction' -v
```

Expected: PASS if alias reuses the same command path; FAIL only if the alias bypasses the new resolver.

- [ ] **Step 3: Fix only if needed**

If the alias fails, wire it through the same options and execution path as `gen`. Do not duplicate selector logic.

- [ ] **Step 4: Run focused command tests**

Run:

```bash
go test ./cmd/vx -run 'TestGenPromptedAction|TestGeneratePromptedAction' -v
```

Expected: PASS.

- [ ] **Step 5: Commit if code or tests changed**

```bash
git add cmd/vx/gen.go cmd/vx/gen_test.go
git commit -m "test: cover prompted generate action"
```

## Chunk 4: Docs And Final Verification

### Task 6: Update User Documentation

**Files:**

- Modify: `README.md`
- Modify: `docs/src/content/docs/index.md`
- Modify: `docs/src/content/docs/commands/gen.md`

- [ ] **Step 1: Update README wording**

Document:

```bash
vx gen vandor/go-backend-core -i
```

Explain:

- `-i, --prompt` fills missing inputs.
- After prompted input, `vx` asks whether to `Preview` or `Apply`.
- `Preview` is the default.
- `vx gen <target> -i --apply` skips the action selector and applies directly.

- [ ] **Step 2: Regenerate docs**

Run:

```bash
just docs-generate
```

Expected: generated docs update from README and command help.

- [ ] **Step 3: Review generated docs**

Run:

```bash
git diff -- README.md docs/src/content/docs/index.md docs/src/content/docs/commands/gen.md
```

Expected: docs mention prompt action selector and no fake `vx prompt` command page appears.

- [ ] **Step 4: Commit docs**

```bash
git add README.md docs/src/content/docs/index.md docs/src/content/docs/commands/gen.md
git commit -m "docs: document prompted generation action"
```

### Task 7: Final Verification

**Files:**

- Verify all changed files.

- [ ] **Step 1: Run focused tests**

Run:

```bash
go test ./internal/ui -v
go test ./cmd/vx -v
```

Expected: PASS.

- [ ] **Step 2: Run full tests**

Run:

```bash
just test
```

Expected: PASS.

- [ ] **Step 3: Run build**

Run:

```bash
just build
```

Expected: PASS.

- [ ] **Step 4: Run diff whitespace check**

Run:

```bash
git diff --check
```

Expected: no output.

- [ ] **Step 5: Docs build note**

Run only if the known `@astrojs/sitemap` issue is being investigated:

```bash
just docs-build
```

Expected today: may fail with the known non-blocking sitemap error `Cannot read properties of undefined (reading 'reduce')` after Astro build/Pagefind. Do not treat that as a regression unless the failure changes.

## Integration Notes

- This plan intentionally does not change the existing prompt-mode design spec. It is a small follow-up behavior on top of `-i, --prompt`.
- If future design work wants a broader wizard flow, write a separate spec. This plan only adds an action selector after a real missing-input prompt session.
- Keep `cmd/vx` thin: command code may decide the final apply boolean, but rendering, planning, and writing remain in existing services.
