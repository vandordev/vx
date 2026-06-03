package input

import "fmt"

type RequiredField struct {
	Name     string
	TypeName string
}

func MissingRequiredInputs(fields []RequiredField, values map[string]any) []RequiredField {
	var missing []RequiredField
	for _, field := range fields {
		if isInjectedProjectField(field.Name) {
			continue
		}
		if !hasValue(values, field.Name) {
			missing = append(missing, field)
		}
	}
	return missing
}

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
	if typeName == "string" {
		return raw, nil
	}

	value, err := parseScalar(raw)
	if err != nil {
		return nil, fmt.Errorf("parse prompted value for %q: %w", name, err)
	}
	return value, nil
}

func isInjectedProjectField(name string) bool {
	return name == "project" || len(name) > len("project.") && name[:len("project.")] == "project."
}

func hasValue(values map[string]any, name string) bool {
	if values == nil {
		return false
	}

	parts := splitDottedPath(name)
	current := values
	for idx, part := range parts {
		value, ok := current[part]
		if !ok {
			return false
		}
		if idx == len(parts)-1 {
			return true
		}
		next, ok := value.(map[string]any)
		if !ok {
			return false
		}
		current = next
	}
	return false
}

func setDottedValue(values map[string]any, name string, value any) {
	parts := splitDottedPath(name)
	current := values
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = value
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneValue(value)
	}
	return cloned
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		cloned := make([]any, len(typed))
		for i, item := range typed {
			cloned[i] = cloneValue(item)
		}
		return cloned
	default:
		return value
	}
}

func splitDottedPath(name string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(name); i++ {
		if name[i] != '.' {
			continue
		}
		parts = append(parts, name[start:i])
		start = i + 1
	}
	parts = append(parts, name[start:])
	return parts
}
