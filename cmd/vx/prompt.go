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

var (
	stdinIsTerminal  = func() bool { return tty.IsTerminal(os.Stdin.Fd()) }
	stdoutIsTerminal = func() bool { return tty.IsTerminal(os.Stdout.Fd()) }
	promptInputs     = realPromptInputs
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

	rawValues, err := promptInputs(toPromptFields(missing))
	if err != nil {
		return nil, err
	}

	prompted := make(map[string]any, len(missing))
	for _, field := range missing {
		rawValue, ok := rawValues[field.Name]
		if !ok {
			return nil, fmt.Errorf("missing prompted value for %q", field.Name)
		}

		value, err := input.ParsePromptValue(field.Name, field.TypeName, rawValue)
		if err != nil {
			return nil, err
		}
		prompted[field.Name] = value
	}

	return input.MergePromptedValues(values, prompted), nil
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
