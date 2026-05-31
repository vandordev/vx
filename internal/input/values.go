package input

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

func Load(valuesPath string, sets []string) (map[string]any, error) {
	values := map[string]any{}

	if strings.TrimSpace(valuesPath) != "" {
		loaded, err := loadValuesFile(valuesPath)
		if err != nil {
			return nil, err
		}
		values = loaded
	}

	for _, setValue := range sets {
		if err := applySet(values, setValue); err != nil {
			return nil, err
		}
	}

	return values, nil
}

func loadValuesFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read values file: %w", err)
	}

	values := map[string]any{}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("parse yaml values file: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("parse json values file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported values file format %q", filepath.Ext(path))
	}

	return values, nil
}

func applySet(values map[string]any, raw string) error {
	key, rawValue, ok := strings.Cut(raw, "=")
	if !ok || strings.TrimSpace(key) == "" {
		return fmt.Errorf("invalid --set value %q", raw)
	}

	parsedValue, err := parseScalar(rawValue)
	if err != nil {
		return fmt.Errorf("parse --set value %q: %w", raw, err)
	}

	parts := strings.Split(key, ".")
	current := values
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = parsedValue
	return nil
}

func parseScalar(raw string) (any, error) {
	var value any
	if err := yaml.Unmarshal([]byte(raw), &value); err != nil {
		return nil, err
	}
	return value, nil
}
