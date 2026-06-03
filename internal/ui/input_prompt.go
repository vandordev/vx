package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

type PromptField struct {
	Name     string
	TypeName string
}

func InputPromptLabels(fields []PromptField) []string {
	labels := make([]string, 0, len(fields))
	for _, field := range fields {
		labels = append(labels, promptLabel(field))
	}
	return labels
}

func PromptInputs(fields []PromptField, theme Theme) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	bindings := make([]promptBinding, 0, len(fields))
	formFields := make([]huh.Field, 0, len(fields))
	for _, field := range fields {
		value := ""
		bindings = append(bindings, promptBinding{
			name:  field.Name,
			value: &value,
		})
		formFields = append(formFields, huh.NewInput().
			Title(promptLabel(field)).
			Prompt("> ").
			Value(&value))
	}

	form := huh.NewForm(huh.NewGroup(formFields...)).
		WithTheme(inputPromptTheme(theme)).
		WithShowHelp(false)
	if err := form.Run(); err != nil {
		return nil, err
	}

	for _, binding := range bindings {
		values[binding.name] = *binding.value
	}
	return values, nil
}

type promptBinding struct {
	name  string
	value *string
}

func promptLabel(field PromptField) string {
	return fmt.Sprintf("%s (%s)", field.Name, field.TypeName)
}

func inputPromptTheme(theme Theme) *huh.Theme {
	huhTheme := huh.ThemeBase()
	huhTheme.Focused.Base = huhTheme.Focused.Base.Foreground(theme.Text)
	huhTheme.Blurred.Base = huhTheme.Blurred.Base.Foreground(theme.Text)
	huhTheme.Focused.Title = huhTheme.Focused.Title.Foreground(theme.Headings).Bold(true)
	huhTheme.Blurred.Title = huhTheme.Blurred.Title.Foreground(theme.Headings).Bold(true)
	huhTheme.Focused.Description = huhTheme.Focused.Description.Foreground(theme.Muted)
	huhTheme.Blurred.Description = huhTheme.Blurred.Description.Foreground(theme.Muted)
	huhTheme.Focused.TextInput.Prompt = huhTheme.Focused.TextInput.Prompt.Foreground(theme.Primary)
	huhTheme.Blurred.TextInput.Prompt = huhTheme.Blurred.TextInput.Prompt.Foreground(theme.Primary)
	huhTheme.Focused.TextInput.Cursor = huhTheme.Focused.TextInput.Cursor.Foreground(theme.Primary)
	huhTheme.Blurred.TextInput.Cursor = huhTheme.Blurred.TextInput.Cursor.Foreground(theme.Primary)
	huhTheme.Focused.TextInput.Placeholder = huhTheme.Focused.TextInput.Placeholder.Foreground(theme.Muted)
	huhTheme.Blurred.TextInput.Placeholder = huhTheme.Blurred.TextInput.Placeholder.Foreground(theme.Muted)
	huhTheme.Focused.TextInput.Text = huhTheme.Focused.TextInput.Text.Foreground(theme.Text)
	huhTheme.Blurred.TextInput.Text = huhTheme.Blurred.TextInput.Text.Foreground(theme.Text)
	huhTheme.Help.ShortKey = huhTheme.Help.ShortKey.Foreground(theme.Muted)
	huhTheme.Help.ShortDesc = huhTheme.Help.ShortDesc.Foreground(theme.Muted)
	return huhTheme
}
