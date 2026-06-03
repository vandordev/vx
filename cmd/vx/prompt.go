package main

import (
	"fmt"
	"os"

	"github.com/vandordev/vx/internal/adapters/tty"
	"github.com/vandordev/vx/internal/domain"
	"github.com/vandordev/vx/internal/input"
	"github.com/vandordev/vx/internal/resolve"
	"github.com/vandordev/vx/internal/ui"
	viewsvc "github.com/vandordev/vx/internal/view"
)

type promptOptions struct {
	Prompt         bool
	AsJSON         bool
	NonInteractive bool
	RequiresPlan   bool
}

type promptedInputResult struct {
	Values   map[string]any
	Prompted bool
}

type promptedGenerationApplyOptions struct {
	ExplicitApply bool
	Prompted      bool
}

var (
	stdinIsTerminal        = func() bool { return tty.IsTerminal(os.Stdin.Fd()) }
	stdoutIsTerminal       = func() bool { return tty.IsTerminal(os.Stdout.Fd()) }
	promptInputs           = realPromptInputs
	promptGenerationAction = realPromptGenerationAction
)

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

func resolvePromptedInput(values map[string]any, fields []input.RequiredField, prompt bool) (promptedInputResult, error) {
	if !prompt {
		return promptedInputResult{Values: values}, nil
	}

	missing := input.MissingRequiredInputs(fields, values)
	if len(missing) == 0 {
		return promptedInputResult{Values: values}, nil
	}
	if !stdinIsTerminal() || !stdoutIsTerminal() {
		return promptedInputResult{}, fmt.Errorf("cannot prompt for input in a non-interactive terminal")
	}

	rawValues, err := promptInputs(toPromptFields(missing))
	if err != nil {
		return promptedInputResult{}, err
	}

	prompted := make(map[string]any, len(missing))
	for _, field := range missing {
		rawValue, ok := rawValues[field.Name]
		if !ok {
			return promptedInputResult{}, fmt.Errorf("missing prompted value for %q", field.Name)
		}

		value, err := input.ParsePromptValue(field.Name, field.TypeName, rawValue)
		if err != nil {
			return promptedInputResult{}, err
		}
		prompted[field.Name] = value
	}

	return promptedInputResult{
		Values:   input.MergePromptedValues(values, prompted),
		Prompted: true,
	}, nil
}

func toPromptFields(fields []input.RequiredField) []ui.PromptField {
	promptFields := make([]ui.PromptField, 0, len(fields))
	for _, field := range fields {
		promptFields = append(promptFields, ui.PromptField{
			Name:     field.Name,
			TypeName: field.TypeName,
		})
	}
	return promptFields
}

func realPromptInputs(fields []ui.PromptField) (map[string]string, error) {
	return ui.PromptInputs(fields, ui.ThemeFromConfig(domain.Config{}))
}

func realPromptGenerationAction() (ui.GenerationAction, error) {
	return ui.PromptGenerationAction(ui.ThemeFromConfig(domain.Config{}))
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

func requiredFieldsForTarget(projectRoot string, target resolve.ResolvedTarget) ([]input.RequiredField, error) {
	result, err := viewsvc.Inspect(viewsvc.Request{
		ProjectRoot: projectRoot,
		Target:      target,
	})
	if err != nil {
		return nil, err
	}

	fields := make([]input.RequiredField, 0, len(result.RequiredInputs))
	for _, field := range result.RequiredInputs {
		fields = append(fields, input.RequiredField{
			Name:     field.Name,
			TypeName: field.TypeName,
		})
	}
	return fields, nil
}
