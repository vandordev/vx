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

func PromptGenerationAction(theme Theme) (GenerationAction, error) {
	action := DefaultGenerationAction()
	options := GenerationActionOptions()

	selectField := huh.NewSelect[GenerationAction]().
		Title("What should vx do next?").
		Inline(true).
		Options(
			huh.NewOption(options[0].Label, options[0].Value),
			huh.NewOption(options[1].Label, options[1].Value),
		).
		Value(&action)

	form := huh.NewForm(huh.NewGroup(selectField)).
		WithTheme(inputPromptTheme(theme)).
		WithShowHelp(false)
	if err := form.Run(); err != nil {
		return "", err
	}

	return action, nil
}
