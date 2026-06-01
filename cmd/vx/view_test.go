package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestViewCommandIsRegistered(t *testing.T) {
	cmd := newRootCmd()

	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("expected view command to be registered, got error: %v", err)
	}
	if viewCmd == nil || viewCmd.Name() != "view" {
		t.Fatalf("expected to resolve view command, got %#v", viewCmd)
	}
}

func TestViewCommandHelp(t *testing.T) {
	cmd := newRootCmd()

	output, err := testutil.RunCLI(t, cmd, "view", "--help")
	if err != nil {
		t.Fatalf("expected view help to succeed, got error: %v\noutput:\n%s", err, output)
	}

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output for view command, output:\n%s", output)
	}
}

func TestViewCommandOutput(t *testing.T) {
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
  doctor:
    kind: task
    steps:
      - type: check-file
        path: go.mod
      - type: format
        tool: gofmt
  handler:
    kind: template
    templates:
      - path: templates/handler.vxt
`,
		filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "usecase.vxt"): `
@template usecase
@input name string
@file "internal/{{ name }}.txt"
hello {{ name }}
@endfile
`,
		filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "handler.vxt"): `
@template handler
@input name string
@file "internal/{{ name }}_handler.txt"
hello {{ name }}
@endfile
`,
		filepath.Join("vpkg", "acme", "api-kit", "vpkg.yaml"): `
apiVersion: vandor.dev/v1alpha1
name: acme/api-kit
version: 0.2.0
kind: template-pack
exports:
  default:
    kind: template
    templates:
      - path: templates/api.vxt
  handler:
    kind: template
    templates:
      - path: templates/handler.vxt
`,
		filepath.Join("vpkg", "acme", "api-kit", "templates", "api.vxt"): `
@template api
@input name string
@file "api/{{ name }}.txt"
hello {{ name }}
@endfile
`,
		filepath.Join("vpkg", "acme", "api-kit", "templates", "handler.vxt"): `
@template handler
@input name string
@file "api/{{ name }}_handler.txt"
hello {{ name }}
@endfile
`,
	})

	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(fixture.Root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})

	t.Run("prints package summary text", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "view", "vandor/go-backend-core")
		if err != nil {
			t.Fatalf("view package summary returned error: %v\noutput:\n%s", err, output)
		}
		for _, snippet := range []string{
			"Package: vandor/go-backend-core",
			"Version: 0.1.0",
			"Exports:",
			"default (template)",
			"doctor (task)",
		} {
			if !strings.Contains(output, snippet) {
				t.Fatalf("expected package summary to contain %q, output:\n%s", snippet, output)
			}
		}
	})

	t.Run("prints export summary text", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "view", "vandor/go-backend-core:default")
		if err != nil {
			t.Fatalf("view export summary returned error: %v\noutput:\n%s", err, output)
		}
		for _, snippet := range []string{
			"Export: default",
			"Kind: template",
			"templates/usecase.vxt",
			"name (string)",
		} {
			if !strings.Contains(output, snippet) {
				t.Fatalf("expected export summary to contain %q, output:\n%s", snippet, output)
			}
		}
	})

	t.Run("prints ordered task steps with note", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "view", "vandor/go-backend-core:doctor")
		if err != nil {
			t.Fatalf("view task summary returned error: %v\noutput:\n%s", err, output)
		}
		for _, snippet := range []string{
			"view-only in v0.1",
			"1. type=check-file path=go.mod",
			"2. type=format tool=gofmt",
		} {
			if !strings.Contains(output, snippet) {
				t.Fatalf("expected task summary to contain %q, output:\n%s", snippet, output)
			}
		}
	})

	t.Run("prints stable json output", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "view", "vandor/go-backend-core:default", "--json")
		if err != nil {
			t.Fatalf("view json returned error: %v\noutput:\n%s", err, output)
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(output), &payload); err != nil {
			t.Fatalf("unmarshal json output: %v\noutput:\n%s", err, output)
		}

		for _, key := range []string{"projectRoot", "sourceClass", "package", "export", "requiredInputs"} {
			if _, ok := payload[key]; !ok {
				t.Fatalf("expected json output to contain key %q, payload: %#v", key, payload)
			}
		}
	})

	t.Run("fails with candidates on ambiguous shorthand", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "view", "handler")
		if err == nil {
			t.Fatalf("expected ambiguous shorthand to fail, output:\n%s", output)
		}
		for _, snippet := range []string{
			"acme/api-kit:handler",
			"vandor/go-backend-core:handler",
		} {
			if !strings.Contains(output, snippet) {
				t.Fatalf("expected ambiguity output to contain %q, output:\n%s", snippet, output)
			}
		}
	})
}

func TestViewCommandInjectsProjectContextForPlan(t *testing.T) {
	fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
		VPKGRoots: []string{"."},
		Directories: []string{
			filepath.Join("services", "billing", "internal", "app"),
		},
	})
	fixture.WriteFiles(t, map[string]string{
		filepath.Join("services", "billing", "go.mod"): "module github.com/acme/platform/services/billing\n\ngo 1.25\n",
		filepath.Join("vpkg", "vandor", "go-backend-core", "vpkg.yaml"): `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: template
    templates:
      - path: templates/context.vxt
`,
		filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "context.vxt"): `
@template context
@input name string
@file "{{ project.go.module_root }}/{{ name }}.txt"
module={{ project.go.module }}
@endfile
`,
	})

	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(fixture.Path("services", "billing", "internal", "app")); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})

	output, err := testutil.RunCLI(t, newRootCmd(), "view", "vandor/go-backend-core:default", "--plan", "--set", "name=booking")
	if err != nil {
		t.Fatalf("view returned error: %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, filepath.Join("services", "billing", "booking.txt")) {
		t.Fatalf("expected planned output to include module-relative path, output:\n%s", output)
	}
}
