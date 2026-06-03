package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/input"
	"github.com/vandordev/vx/internal/ui"
)

func TestValidatePromptOptions(t *testing.T) {
	t.Run("rejects json", func(t *testing.T) {
		err := validatePromptOptions(promptOptions{
			Prompt:       true,
			AsJSON:       true,
			RequiresPlan: true,
		})
		if err == nil || !strings.Contains(err.Error(), "--prompt cannot be used with --json") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects non interactive", func(t *testing.T) {
		err := validatePromptOptions(promptOptions{
			Prompt:         true,
			NonInteractive: true,
			RequiresPlan:   true,
		})
		if err == nil || !strings.Contains(err.Error(), "--prompt cannot be used with --non-interactive") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects prompt without planning", func(t *testing.T) {
		err := validatePromptOptions(promptOptions{
			Prompt:       true,
			RequiresPlan: false,
		})
		if err == nil || !strings.Contains(err.Error(), "--prompt requires template planning") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("allows non prompt mode", func(t *testing.T) {
		if err := validatePromptOptions(promptOptions{}); err != nil {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestResolvePromptedInput(t *testing.T) {
	t.Run("prompts only missing fields", func(t *testing.T) {
		restoreTTY := stubTTY(t, true, true)
		defer restoreTTY()

		var requested []ui.PromptField
		restorePrompt := stubPrompt(t, func(fields []ui.PromptField) (map[string]string, error) {
			requested = append(requested, fields...)
			return map[string]string{"name": "booking"}, nil
		})
		defer restorePrompt()

		result, err := resolvePromptedInput(map[string]any{"count": 2}, []input.RequiredField{
			{Name: "name", TypeName: "string"},
			{Name: "count", TypeName: "int"},
		}, true)
		if err != nil {
			t.Fatalf("resolvePromptedInput returned error: %v", err)
		}
		values := result.Values
		if values["name"] != "booking" || values["count"] != 2 {
			t.Fatalf("values = %#v", values)
		}
		if !result.Prompted {
			t.Fatalf("Prompted = false, want true")
		}
		if len(requested) != 1 || requested[0].Name != "name" {
			t.Fatalf("requested = %#v", requested)
		}
	})

	t.Run("does not prompt when nothing missing", func(t *testing.T) {
		restoreTTY := stubTTY(t, true, true)
		defer restoreTTY()

		called := 0
		restorePrompt := stubPrompt(t, func(fields []ui.PromptField) (map[string]string, error) {
			called++
			return nil, nil
		})
		defer restorePrompt()

		result, err := resolvePromptedInput(map[string]any{"name": "booking"}, []input.RequiredField{
			{Name: "name", TypeName: "string"},
		}, true)
		if err != nil {
			t.Fatalf("resolvePromptedInput returned error: %v", err)
		}
		values := result.Values
		if values["name"] != "booking" {
			t.Fatalf("values = %#v", values)
		}
		if result.Prompted {
			t.Fatalf("Prompted = true, want false")
		}
		if called != 0 {
			t.Fatalf("prompt called %d times", called)
		}
	})

	t.Run("rejects non tty", func(t *testing.T) {
		restoreTTY := stubTTY(t, false, true)
		defer restoreTTY()

		_, err := resolvePromptedInput(map[string]any{}, []input.RequiredField{
			{Name: "name", TypeName: "string"},
		}, true)
		if err == nil || !strings.Contains(err.Error(), "cannot prompt for input in a non-interactive terminal") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("returns parse errors with input name", func(t *testing.T) {
		restoreTTY := stubTTY(t, true, true)
		defer restoreTTY()

		restorePrompt := stubPrompt(t, func(fields []ui.PromptField) (map[string]string, error) {
			return map[string]string{"enabled": "["}, nil
		})
		defer restorePrompt()

		_, err := resolvePromptedInput(map[string]any{}, []input.RequiredField{
			{Name: "enabled", TypeName: "bool"},
		}, true)
		if err == nil || !strings.Contains(err.Error(), "enabled") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("never prompts injected project fields", func(t *testing.T) {
		restoreTTY := stubTTY(t, true, true)
		defer restoreTTY()

		called := 0
		restorePrompt := stubPrompt(t, func(fields []ui.PromptField) (map[string]string, error) {
			called++
			return map[string]string{}, nil
		})
		defer restorePrompt()

		result, err := resolvePromptedInput(map[string]any{}, []input.RequiredField{
			{Name: "project", TypeName: "object"},
			{Name: "project.go.module", TypeName: "string"},
		}, true)
		if err != nil {
			t.Fatalf("resolvePromptedInput returned error: %v", err)
		}
		values := result.Values
		if len(values) != 0 {
			t.Fatalf("values = %#v", values)
		}
		if result.Prompted {
			t.Fatalf("Prompted = true, want false")
		}
		if called != 0 {
			t.Fatalf("prompt called %d times", called)
		}
	})
}

func stubTTY(t *testing.T, stdinTTY, stdoutTTY bool) func() {
	t.Helper()
	originalStdin := stdinIsTerminal
	originalStdout := stdoutIsTerminal
	stdinIsTerminal = func() bool { return stdinTTY }
	stdoutIsTerminal = func() bool { return stdoutTTY }
	return func() {
		stdinIsTerminal = originalStdin
		stdoutIsTerminal = originalStdout
	}
}

func stubPrompt(t *testing.T, fn func(fields []ui.PromptField) (map[string]string, error)) func() {
	t.Helper()
	original := promptInputs
	promptInputs = fn
	return func() {
		promptInputs = original
	}
}

func TestToPromptFields(t *testing.T) {
	fields := toPromptFields([]input.RequiredField{
		{Name: "name", TypeName: "string"},
		{Name: "enabled", TypeName: "bool"},
	})
	if len(fields) != 2 {
		t.Fatalf("fields = %#v", fields)
	}
	if fields[0].Name != "name" || fields[1].TypeName != "bool" {
		t.Fatalf("fields = %#v", fields)
	}
}

func TestResolvePromptedInputPropagatesPromptError(t *testing.T) {
	restoreTTY := stubTTY(t, true, true)
	defer restoreTTY()

	restorePrompt := stubPrompt(t, func(fields []ui.PromptField) (map[string]string, error) {
		return nil, fmt.Errorf("prompt cancelled")
	})
	defer restorePrompt()

	_, err := resolvePromptedInput(map[string]any{}, []input.RequiredField{
		{Name: "name", TypeName: "string"},
	}, true)
	if err == nil || !strings.Contains(err.Error(), "prompt cancelled") {
		t.Fatalf("error = %v", err)
	}
}

func TestResolvePromptedGenerationApply(t *testing.T) {
	t.Run("uses explicit apply without selector", func(t *testing.T) {
		called := 0
		restore := stubPromptGenerationAction(t, func() (ui.GenerationAction, error) {
			called++
			return ui.GenerationActionPreview, nil
		})
		defer restore()

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
		if called != 0 {
			t.Fatalf("selector called %d times", called)
		}
	})

	t.Run("skips selector when nothing was prompted", func(t *testing.T) {
		called := 0
		restore := stubPromptGenerationAction(t, func() (ui.GenerationAction, error) {
			called++
			return ui.GenerationActionApply, nil
		})
		defer restore()

		apply, err := resolvePromptedGenerationApply(promptedGenerationApplyOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if apply {
			t.Fatalf("apply = true, want false")
		}
		if called != 0 {
			t.Fatalf("selector called %d times", called)
		}
	})

	t.Run("uses preview selection", func(t *testing.T) {
		restore := stubPromptGenerationAction(t, func() (ui.GenerationAction, error) {
			return ui.GenerationActionPreview, nil
		})
		defer restore()

		apply, err := resolvePromptedGenerationApply(promptedGenerationApplyOptions{Prompted: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if apply {
			t.Fatalf("apply = true, want false")
		}
	})

	t.Run("uses apply selection", func(t *testing.T) {
		restore := stubPromptGenerationAction(t, func() (ui.GenerationAction, error) {
			return ui.GenerationActionApply, nil
		})
		defer restore()

		apply, err := resolvePromptedGenerationApply(promptedGenerationApplyOptions{Prompted: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !apply {
			t.Fatalf("apply = false, want true")
		}
	})

	t.Run("propagates selector error", func(t *testing.T) {
		restore := stubPromptGenerationAction(t, func() (ui.GenerationAction, error) {
			return "", fmt.Errorf("selection cancelled")
		})
		defer restore()

		_, err := resolvePromptedGenerationApply(promptedGenerationApplyOptions{Prompted: true})
		if err == nil || !strings.Contains(err.Error(), "selection cancelled") {
			t.Fatalf("error = %v", err)
		}
	})
}

func stubPromptGenerationAction(t *testing.T, fn func() (ui.GenerationAction, error)) func() {
	t.Helper()
	original := promptGenerationAction
	promptGenerationAction = fn
	return func() {
		promptGenerationAction = original
	}
}
