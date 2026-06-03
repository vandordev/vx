package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/project"
	"github.com/vandordev/vx/internal/resolve"
	"github.com/vandordev/vx/internal/testutil"
	"github.com/vandordev/vx/internal/vpkg"
)

func TestGenerate(t *testing.T) {
	fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
		VPKGRoots: []string{"."},
	})
	fixture.WriteFiles(t, map[string]string{
		filepath.Join("vpkg", "vandor", "go-backend-core", "vpkg.yaml"): `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: template
    templates:
      - path: templates/usecase.vxt
  starter:
    kind: preset
    templates:
      - path: templates/usecase.vxt
  doctor:
    kind: task
    steps:
      - type: check-file
        path: go.mod
`,
		filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "usecase.vxt"): `
@template usecase
@input name string
@dir "internal"
@file "internal/{{ name }}.txt"
hello {{ name }}
@endfile
`,
		filepath.Join("templates", "standalone.vxt"): `
@template standalone
@input context string
@file "contexts/{{ context }}.txt"
hello {{ context }}
@endfile
`,
	})

	packages, err := vpkg.Discover(fixture.Root)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}

	t.Run("previews template export", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:default", packages)

		result, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Input:       map[string]any{"name": "booking"},
		})
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if len(result.PlannedDirs) != 1 || result.PlannedDirs[0] != "internal" {
			t.Fatalf("PlannedDirs = %#v", result.PlannedDirs)
		}
		if len(result.PlannedFiles) != 1 || result.PlannedFiles[0].Path != "internal/booking.txt" {
			t.Fatalf("PlannedFiles = %#v", result.PlannedFiles)
		}
		if result.Applied {
			t.Fatal("expected preview mode to avoid writes")
		}
	})

	t.Run("previews direct vxt path", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "./templates/standalone.vxt", packages)

		result, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Input:       map[string]any{"context": "payments"},
		})
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if len(result.PlannedFiles) != 1 || result.PlannedFiles[0].Path != "contexts/payments.txt" {
			t.Fatalf("PlannedFiles = %#v", result.PlannedFiles)
		}
	})

	t.Run("rejects preset export generation", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:starter", packages)

		_, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Input:       map[string]any{"name": "booking"},
		})
		if err == nil || !strings.Contains(err.Error(), "preset") {
			t.Fatalf("Generate error = %v, want preset rejection", err)
		}
	})

	t.Run("rejects task export generation", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:doctor", packages)

		_, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
		})
		if err == nil || !strings.Contains(err.Error(), "task") {
			t.Fatalf("Generate error = %v, want task rejection", err)
		}
	})

	t.Run("applies writes into project root", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:default", packages)

		result, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Input:       map[string]any{"name": "checkout"},
			Apply:       true,
		})
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if !result.Applied {
			t.Fatal("expected apply mode to write files")
		}

		content, err := os.ReadFile(fixture.Path("internal", "checkout.txt"))
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		if !strings.Contains(string(content), "hello checkout") {
			t.Fatalf("generated file content = %q", string(content))
		}
	})

	t.Run("fails apply when target file already exists", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:default", packages)
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("internal", "existing.txt"): "keep me\n",
		})

		_, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Input:       map[string]any{"name": "existing"},
			Apply:       true,
		})
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("Generate error = %v, want conflict error", err)
		}
	})

	t.Run("fails non interactive generation when required input is missing", func(t *testing.T) {
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:default", packages)

		_, err := Generate(Request{
			ProjectRoot:    fixture.Root,
			Target:         target,
			Input:          map[string]any{},
			NonInteractive: true,
		})
		if err == nil || !strings.Contains(err.Error(), "missing input") {
			t.Fatalf("Generate error = %v, want missing input error", err)
		}
	})

	t.Run("injects project context into template input", func(t *testing.T) {
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "context.vxt"): `
@template context
@input name string
@file "internal/{{ name }}_context.txt"
module={{ project.go.module }}
module_root={{ project.go.module_root }}
root={{ project.root }}
language={{ project.language }}
@endfile
`,
			filepath.Join("vpkg", "vandor", "go-backend-core", "vpkg.yaml"): `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  context:
    kind: template
    templates:
      - path: templates/context.vxt
`,
		})
		packages, err := vpkg.Discover(fixture.Root)
		if err != nil {
			t.Fatalf("Discover returned error: %v", err)
		}
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:context", packages)

		_, err = Generate(Request{
			ProjectRoot: fixture.Root,
			ProjectContext: project.Context{
				Root:     fixture.Root,
				Language: "go",
				Go: &project.GoContext{
					Module:     "github.com/acme/platform/services/billing",
					ModuleRoot: filepath.Join("services", "billing"),
				},
			},
			Target: target,
			Input:  map[string]any{"name": "booking"},
			Apply:  true,
		})
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}

		content, err := os.ReadFile(fixture.Path("internal", "booking_context.txt"))
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		for _, snippet := range []string{
			"module=github.com/acme/platform/services/billing",
			"module_root=" + filepath.Join("services", "billing"),
			"root=" + fixture.Root,
			"language=go",
		} {
			if !strings.Contains(string(content), snippet) {
				t.Fatalf("expected generated content to contain %q, content:\n%s", snippet, content)
			}
		}
	})

	t.Run("applies vxt case filters to paths and written content", func(t *testing.T) {
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "filters.vxt"): `
@template filters
@input name string
@file "internal/{{ name | snake }}/{{ name | kebab }}.go"
package {{ name | snake }}

type {{ name | pascal }}Service struct {
	value string // {{ name | camel }}
}
@endfile
`,
			filepath.Join("vpkg", "vandor", "go-backend-core", "vpkg.yaml"): `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  filters:
    kind: template
    templates:
      - path: templates/filters.vxt
`,
		})
		packages, err := vpkg.Discover(fixture.Root)
		if err != nil {
			t.Fatalf("Discover returned error: %v", err)
		}
		target := mustResolveGenerate(t, fixture.Root, "vandor/go-backend-core:filters", packages)

		result, err := Generate(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Input:       map[string]any{"name": "order item"},
			Apply:       true,
		})
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if len(result.PlannedFiles) != 1 {
			t.Fatalf("PlannedFiles = %#v", result.PlannedFiles)
		}
		planned := result.PlannedFiles[0]
		if planned.Path != filepath.Join("internal", "order_item", "order-item.go") {
			t.Fatalf("PlannedFile.Path = %q", planned.Path)
		}

		content, err := os.ReadFile(fixture.Path("internal", "order_item", "order-item.go"))
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		for _, snippet := range []string{
			"package order_item",
			"type OrderItemService struct",
			"value string // orderItem",
		} {
			if !strings.Contains(string(content), snippet) {
				t.Fatalf("expected generated content to contain %q, content:\n%s", snippet, content)
			}
		}
	})
}

func mustResolveGenerate(t *testing.T, projectRoot string, target string, packages []vpkg.Package) resolve.ResolvedTarget {
	t.Helper()

	resolved, err := resolve.Resolve(projectRoot, target, packages, resolve.ModeGenerate)
	if err != nil {
		t.Fatalf("Resolve(%q) returned error: %v", target, err)
	}
	return resolved
}
