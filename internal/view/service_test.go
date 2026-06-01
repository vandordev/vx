package view

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/project"
	"github.com/vandordev/vx/internal/resolve"
	"github.com/vandordev/vx/internal/testutil"
	"github.com/vandordev/vx/internal/vpkg"
)

func TestInspect(t *testing.T) {
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
    files:
      - path: README.md
    directories:
      - path: internal/app
  doctor:
    kind: task
    steps:
      - type: check-file
        path: go.mod
      - type: format
        tool: gofmt
`,
		filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "usecase.vxt"): `
@template usecase
@input name string
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

	t.Run("returns package summary", func(t *testing.T) {
		target := mustResolve(t, fixture.Root, "vandor/go-backend-core", packages, resolve.ModeView)

		result, err := Inspect(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
		})
		if err != nil {
			t.Fatalf("Inspect returned error: %v", err)
		}
		if result.Package.Name != "vandor/go-backend-core" {
			t.Fatalf("Package.Name = %q", result.Package.Name)
		}
		if len(result.Package.Exports) != 3 {
			t.Fatalf("got %d exports", len(result.Package.Exports))
		}
		if result.Export.Name != "" {
			t.Fatalf("expected no selected export for package summary, got %q", result.Export.Name)
		}
	})

	t.Run("returns template export summary with required inputs", func(t *testing.T) {
		target := mustResolve(t, fixture.Root, "vandor/go-backend-core:default", packages, resolve.ModeView)

		result, err := Inspect(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
		})
		if err != nil {
			t.Fatalf("Inspect returned error: %v", err)
		}
		if result.Export.Kind != "template" {
			t.Fatalf("Export.Kind = %q", result.Export.Kind)
		}
		if len(result.Export.TemplatePaths) != 1 || result.Export.TemplatePaths[0] != "templates/usecase.vxt" {
			t.Fatalf("TemplatePaths = %#v", result.Export.TemplatePaths)
		}
		if len(result.RequiredInputs) != 1 || result.RequiredInputs[0].Name != "name" || result.RequiredInputs[0].TypeName != "string" {
			t.Fatalf("RequiredInputs = %#v", result.RequiredInputs)
		}
	})

	t.Run("returns preset export summary", func(t *testing.T) {
		target := mustResolve(t, fixture.Root, "vandor/go-backend-core:starter", packages, resolve.ModeView)

		result, err := Inspect(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
		})
		if err != nil {
			t.Fatalf("Inspect returned error: %v", err)
		}
		if result.Export.Kind != "preset" {
			t.Fatalf("Export.Kind = %q", result.Export.Kind)
		}
		if len(result.Export.Files) != 1 || result.Export.Files[0] != "README.md" {
			t.Fatalf("Files = %#v", result.Export.Files)
		}
		if len(result.Export.Directories) != 1 || result.Export.Directories[0] != "internal/app" {
			t.Fatalf("Directories = %#v", result.Export.Directories)
		}
	})

	t.Run("returns task export summary with view only note", func(t *testing.T) {
		target := mustResolve(t, fixture.Root, "vandor/go-backend-core:doctor", packages, resolve.ModeView)

		result, err := Inspect(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
		})
		if err != nil {
			t.Fatalf("Inspect returned error: %v", err)
		}
		if result.Export.Kind != "task" {
			t.Fatalf("Export.Kind = %q", result.Export.Kind)
		}
		if len(result.Export.Steps) != 2 {
			t.Fatalf("Steps = %#v", result.Export.Steps)
		}
		if !strings.Contains(result.Export.Note, "view-only in v0.1") {
			t.Fatalf("Note = %q", result.Export.Note)
		}
	})

	t.Run("returns direct vxt summary", func(t *testing.T) {
		target := mustResolve(t, fixture.Root, "./templates/standalone.vxt", packages, resolve.ModeView)

		result, err := Inspect(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
		})
		if err != nil {
			t.Fatalf("Inspect returned error: %v", err)
		}
		if result.SourceClass != resolve.SourceClassDirectVXT {
			t.Fatalf("SourceClass = %q", result.SourceClass)
		}
		if len(result.RequiredInputs) != 1 || result.RequiredInputs[0].Name != "context" {
			t.Fatalf("RequiredInputs = %#v", result.RequiredInputs)
		}
	})

	t.Run("fails clearly when plan input is incomplete", func(t *testing.T) {
		target := mustResolve(t, fixture.Root, "vandor/go-backend-core:default", packages, resolve.ModeView)

		_, err := Inspect(Request{
			ProjectRoot: fixture.Root,
			Target:      target,
			Plan:        true,
			Input:       map[string]any{},
		})
		if err == nil || !strings.Contains(err.Error(), "missing input") {
			t.Fatalf("Inspect error = %v, want missing input error", err)
		}
	})

	t.Run("injects project context into planned template input", func(t *testing.T) {
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "context.vxt"): `
@template context
@input name string
@file "{{ project.go.module_root }}/{{ name }}.txt"
module {{ project.go.module }}
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
		target := mustResolve(t, fixture.Root, "vandor/go-backend-core:context", packages, resolve.ModeView)

		result, err := Inspect(Request{
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
			Plan:   true,
		})
		if err != nil {
			t.Fatalf("Inspect returned error: %v", err)
		}
		if len(result.PlannedFiles) != 1 || result.PlannedFiles[0].Path != filepath.Join("services", "billing", "booking.txt") {
			t.Fatalf("PlannedFiles = %#v", result.PlannedFiles)
		}
	})
}

func mustResolve(t *testing.T, projectRoot string, target string, packages []vpkg.Package, mode resolve.Mode) resolve.ResolvedTarget {
	t.Helper()

	resolved, err := resolve.Resolve(projectRoot, target, packages, mode)
	if err != nil {
		t.Fatalf("Resolve(%q) returned error: %v", target, err)
	}
	return resolved
}
