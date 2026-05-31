package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("loads yaml values file", func(t *testing.T) {
		valuesPath := writeValuesFile(t, "values.yaml", `
context: booking
feature:
  name: create_booking
`)

		values, err := Load(valuesPath, nil)
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}
		if values["context"] != "booking" {
			t.Fatalf("context = %#v", values["context"])
		}
		feature, ok := values["feature"].(map[string]any)
		if !ok || feature["name"] != "create_booking" {
			t.Fatalf("feature = %#v", values["feature"])
		}
	})

	t.Run("loads json values file", func(t *testing.T) {
		valuesPath := writeValuesFile(t, "values.json", `{"context":"booking","count":2}`)

		values, err := Load(valuesPath, nil)
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}
		if values["context"] != "booking" {
			t.Fatalf("context = %#v", values["context"])
		}
		if values["count"] != float64(2) {
			t.Fatalf("count = %#v", values["count"])
		}
	})

	t.Run("set overrides file values", func(t *testing.T) {
		valuesPath := writeValuesFile(t, "values.yaml", `
context: booking
name: old_name
`)

		values, err := Load(valuesPath, []string{"name=new_name"})
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}
		if values["name"] != "new_name" {
			t.Fatalf("name = %#v", values["name"])
		}
	})

	t.Run("set supports dotted paths", func(t *testing.T) {
		values, err := Load("", []string{
			"context=booking",
			"feature.name=create_booking",
			"feature.enabled=true",
		})
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		feature, ok := values["feature"].(map[string]any)
		if !ok {
			t.Fatalf("feature = %#v", values["feature"])
		}
		if feature["name"] != "create_booking" {
			t.Fatalf("feature.name = %#v", feature["name"])
		}
		if feature["enabled"] != true {
			t.Fatalf("feature.enabled = %#v", feature["enabled"])
		}
	})

	t.Run("fails on invalid values file format", func(t *testing.T) {
		valuesPath := writeValuesFile(t, "values.txt", "context=booking\n")

		_, err := Load(valuesPath, nil)
		if err == nil {
			t.Fatal("expected Load to fail for unsupported values format")
		}
	})
}

func writeValuesFile(t *testing.T, name, content string) string {
	t.Helper()

	root := t.TempDir()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write values file: %v", err)
	}
	return path
}
