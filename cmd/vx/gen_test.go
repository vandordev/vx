package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestGenCommandIsRegistered(t *testing.T) {
	cmd := newRootCmd()

	genCmd, _, err := cmd.Find([]string{"gen"})
	if err != nil {
		t.Fatalf("expected gen command to be registered, got error: %v", err)
	}
	if genCmd == nil || genCmd.Name() != "gen" {
		t.Fatalf("expected to resolve gen command, got %#v", genCmd)
	}
}

func TestGenerateAliasResolvesToGen(t *testing.T) {
	cmd := newRootCmd()

	genCmd, _, err := cmd.Find([]string{"gen"})
	if err != nil {
		t.Fatalf("expected gen command to be registered, got error: %v", err)
	}

	aliasCmd, _, err := cmd.Find([]string{"generate"})
	if err != nil {
		t.Fatalf("expected generate alias to resolve, got error: %v", err)
	}
	if aliasCmd == nil || aliasCmd.Name() != "gen" {
		t.Fatalf("expected generate alias to resolve to gen, got %#v", aliasCmd)
	}
	if aliasCmd != genCmd {
		t.Fatalf("expected generate alias to resolve to the gen command instance")
	}
}

func TestGenCommandHelp(t *testing.T) {
	cmd := newRootCmd()

	output, err := testutil.RunCLI(t, cmd, "gen", "--help")
	if err != nil {
		t.Fatalf("expected gen help to succeed, got error: %v\noutput:\n%s", err, output)
	}

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output for gen command, output:\n%s", output)
	}
}

func TestGenerateAliasHelp(t *testing.T) {
	cmd := newRootCmd()

	output, err := testutil.RunCLI(t, cmd, "generate", "--help")
	if err != nil {
		t.Fatalf("expected generate alias help to succeed, got error: %v\noutput:\n%s", err, output)
	}

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output for generate alias, output:\n%s", output)
	}
}

func TestGenCommandOutput(t *testing.T) {
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
`,
		filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "usecase.vxt"): `
@template usecase
@input name string
@dir "internal"
@file "internal/{{ name }}.txt"
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

	t.Run("prints preview output text", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "--set", "name=booking")
		if err != nil {
			t.Fatalf("gen preview returned error: %v\noutput:\n%s", err, output)
		}
		for _, snippet := range []string{
			"Preview: vandor/go-backend-core",
			"Planned Directories:",
			"internal",
			"internal/booking.txt",
		} {
			if !strings.Contains(output, snippet) {
				t.Fatalf("expected preview output to contain %q, output:\n%s", snippet, output)
			}
		}
	})

	t.Run("apply writes files", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "--set", "name=checkout", "--apply")
		if err != nil {
			t.Fatalf("gen apply returned error: %v\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Applied: true") {
			t.Fatalf("expected apply output to confirm write, output:\n%s", output)
		}
		content, err := os.ReadFile(fixture.Path("internal", "checkout.txt"))
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		if !strings.Contains(string(content), "hello checkout") {
			t.Fatalf("generated file content = %q", string(content))
		}
	})

	t.Run("generate alias matches gen json output", func(t *testing.T) {
		genOutput, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "--set", "name=alias", "--json")
		if err != nil {
			t.Fatalf("gen json returned error: %v\noutput:\n%s", err, genOutput)
		}
		aliasOutput, err := testutil.RunCLI(t, newRootCmd(), "generate", "vandor/go-backend-core", "--set", "name=alias", "--json")
		if err != nil {
			t.Fatalf("generate alias json returned error: %v\noutput:\n%s", err, aliasOutput)
		}
		if genOutput != aliasOutput {
			t.Fatalf("expected gen and generate alias to match\nGEN:\n%s\nALIAS:\n%s", genOutput, aliasOutput)
		}
	})

	t.Run("prints stable json result", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "--set", "name=json_case", "--json")
		if err != nil {
			t.Fatalf("gen json returned error: %v\noutput:\n%s", err, output)
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(output), &payload); err != nil {
			t.Fatalf("unmarshal json output: %v\noutput:\n%s", err, output)
		}
		for _, key := range []string{"projectRoot", "sourceClass", "plannedFiles", "applied"} {
			if _, ok := payload[key]; !ok {
				t.Fatalf("expected json output to contain key %q, payload: %#v", key, payload)
			}
		}
	})

	t.Run("returns conflict error without partial success wording", func(t *testing.T) {
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("internal", "existing.txt"): "keep\n",
		})

		output, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "--set", "name=existing", "--apply")
		if err == nil {
			t.Fatalf("expected conflict to fail, output:\n%s", output)
		}
		if !strings.Contains(output, "already exists") {
			t.Fatalf("expected conflict output to mention existing file, output:\n%s", output)
		}
		if strings.Contains(output, "Applied: true") {
			t.Fatalf("expected conflict output to avoid success wording, output:\n%s", output)
		}
	})
}

func TestGenCommandInjectsProjectContext(t *testing.T) {
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
@file "internal/{{ name }}_context.txt"
module={{ project.go.module }}
module_root={{ project.go.module_root }}
@endfile
`,
		filepath.Join("templates", "direct.vxt"): `
@template direct
@file "direct_context.txt"
module={{ project.go.module }}
module_root={{ project.go.module_root }}
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

	t.Run("package export", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "gen", "vandor/go-backend-core", "--set", "name=booking", "--apply")
		if err != nil {
			t.Fatalf("gen returned error: %v\noutput:\n%s", err, output)
		}

		assertGeneratedProjectContext(t, fixture.Path("internal", "booking_context.txt"))
	})

	t.Run("direct vxt source", func(t *testing.T) {
		output, err := testutil.RunCLI(t, newRootCmd(), "gen", "./templates/direct.vxt", "--apply")
		if err != nil {
			t.Fatalf("gen direct returned error: %v\noutput:\n%s", err, output)
		}

		assertGeneratedProjectContext(t, fixture.Path("direct_context.txt"))
	})
}

func assertGeneratedProjectContext(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	for _, snippet := range []string{
		"module=github.com/acme/platform/services/billing",
		"module_root=" + filepath.Join("services", "billing"),
	} {
		if !strings.Contains(string(content), snippet) {
			t.Fatalf("expected generated content to contain %q, content:\n%s", snippet, content)
		}
	}
}
