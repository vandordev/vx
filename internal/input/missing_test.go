package input_test

import (
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/input"
)

func TestMissingRequiredInputs(t *testing.T) {
	t.Run("all required values present", func(t *testing.T) {
		fields := []input.RequiredField{
			{Name: "name", TypeName: "string"},
			{Name: "feature.enabled", TypeName: "bool"},
			{Name: "empty", TypeName: "string"},
			{Name: "count", TypeName: "int"},
			{Name: "project", TypeName: "object"},
		}
		values := map[string]any{
			"name":  "booking",
			"empty": "",
			"count": 0,
			"feature": map[string]any{
				"enabled": false,
			},
		}

		missing := input.MissingRequiredInputs(fields, values)
		if len(missing) != 0 {
			t.Fatalf("missing = %#v", missing)
		}
	})

	t.Run("reports missing top level value", func(t *testing.T) {
		fields := []input.RequiredField{{Name: "name", TypeName: "string"}}

		missing := input.MissingRequiredInputs(fields, map[string]any{})
		if len(missing) != 1 || missing[0].Name != "name" {
			t.Fatalf("missing = %#v", missing)
		}
	})

	t.Run("reports missing dotted value when parent map absent", func(t *testing.T) {
		fields := []input.RequiredField{{Name: "feature.enabled", TypeName: "bool"}}

		missing := input.MissingRequiredInputs(fields, map[string]any{})
		if len(missing) != 1 || missing[0].Name != "feature.enabled" {
			t.Fatalf("missing = %#v", missing)
		}
	})

	t.Run("reports missing dotted value when leaf absent", func(t *testing.T) {
		fields := []input.RequiredField{{Name: "feature.enabled", TypeName: "bool"}}
		values := map[string]any{
			"feature": map[string]any{},
		}

		missing := input.MissingRequiredInputs(fields, values)
		if len(missing) != 1 || missing[0].Name != "feature.enabled" {
			t.Fatalf("missing = %#v", missing)
		}
	})

	t.Run("skips injected project fields", func(t *testing.T) {
		fields := []input.RequiredField{
			{Name: "project", TypeName: "object"},
			{Name: "project.go.module", TypeName: "string"},
		}

		missing := input.MissingRequiredInputs(fields, map[string]any{})
		if len(missing) != 0 {
			t.Fatalf("missing = %#v", missing)
		}
	})
}

func TestMergePromptedValues(t *testing.T) {
	t.Run("does not overwrite existing values", func(t *testing.T) {
		values := map[string]any{"name": "from-set"}
		prompted := map[string]any{"name": "from-prompt", "count": 2}

		merged := input.MergePromptedValues(values, prompted)

		if merged["name"] != "from-set" {
			t.Fatalf("name = %#v", merged["name"])
		}
		if merged["count"] != 2 {
			t.Fatalf("count = %#v", merged["count"])
		}
		if values["count"] != nil {
			t.Fatalf("input map mutated = %#v", values)
		}
	})

	t.Run("merges dotted prompted keys", func(t *testing.T) {
		merged := input.MergePromptedValues(map[string]any{}, map[string]any{
			"feature.enabled": true,
		})

		feature, ok := merged["feature"].(map[string]any)
		if !ok {
			t.Fatalf("feature = %#v", merged["feature"])
		}
		if feature["enabled"] != true {
			t.Fatalf("feature.enabled = %#v", feature["enabled"])
		}
	})
}

func TestParsePromptValue(t *testing.T) {
	t.Run("string stays string", func(t *testing.T) {
		value, err := input.ParsePromptValue("name", "string", "123")
		if err != nil {
			t.Fatalf("ParsePromptValue returned error: %v", err)
		}
		if value != "123" {
			t.Fatalf("value = %#v", value)
		}
	})

	t.Run("bool uses scalar parsing", func(t *testing.T) {
		value, err := input.ParsePromptValue("enabled", "bool", "true")
		if err != nil {
			t.Fatalf("ParsePromptValue returned error: %v", err)
		}
		if value != true {
			t.Fatalf("value = %#v", value)
		}
	})

	t.Run("numeric uses scalar parsing", func(t *testing.T) {
		value, err := input.ParsePromptValue("count", "int", "2")
		if err != nil {
			t.Fatalf("ParsePromptValue returned error: %v", err)
		}
		if value != 2 {
			t.Fatalf("value = %#v", value)
		}
	})

	t.Run("invalid bool returns input name", func(t *testing.T) {
		_, err := input.ParsePromptValue("enabled", "bool", "[")
		if err == nil || !strings.Contains(err.Error(), "enabled") {
			t.Fatalf("error = %v", err)
		}
	})
}
