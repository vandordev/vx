package resolve

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
	"github.com/vandordev/vx/internal/vpkg"
)

func TestResolve(t *testing.T) {
	fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
		VPKGRoots: []string{"."},
	})
	fixture.WriteFiles(t, map[string]string{
		filepath.Join("templates", "direct.vxt"): "name: direct\n",
	})

	packages := []vpkg.Package{
		{
			Path: "vandor/go-backend-core",
			Manifest: vpkg.Manifest{
				Name: "vandor/go-backend-core",
				Exports: map[string]vpkg.Export{
					"default": {Kind: "template"},
					"usecase": {Kind: "template"},
				},
			},
		},
		{
			Path: "acme/api-kit",
			Manifest: vpkg.Manifest{
				Name: "acme/api-kit",
				Exports: map[string]vpkg.Export{
					"default": {Kind: "template"},
					"handler": {Kind: "template"},
				},
			},
		},
	}

	t.Run("resolves direct vxt path inside project root", func(t *testing.T) {
		target, err := Resolve(fixture.Root, "./templates/direct.vxt", packages, ModeView)
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if target.SourceClass != SourceClassDirectVXT {
			t.Fatalf("SourceClass = %q, want %q", target.SourceClass, SourceClassDirectVXT)
		}
		if target.TemplatePath != fixture.Path("templates", "direct.vxt") {
			t.Fatalf("TemplatePath = %q", target.TemplatePath)
		}
	})

	t.Run("rejects direct vxt path outside project root", func(t *testing.T) {
		outside := filepath.Join(t.TempDir(), "outside.vxt")
		if err := os.WriteFile(outside, []byte("name: outside\n"), 0o644); err != nil {
			t.Fatalf("write outside template: %v", err)
		}

		_, err := Resolve(fixture.Root, outside, packages, ModeView)
		if err == nil || !strings.Contains(err.Error(), "project root") {
			t.Fatalf("Resolve error = %v, want project root boundary error", err)
		}
	})

	t.Run("resolves full package reference in view mode", func(t *testing.T) {
		target, err := Resolve(fixture.Root, "vandor/go-backend-core", packages, ModeView)
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if target.SourceClass != SourceClassPackage {
			t.Fatalf("SourceClass = %q, want %q", target.SourceClass, SourceClassPackage)
		}
		if target.Package.Path != "vandor/go-backend-core" {
			t.Fatalf("Package.Path = %q", target.Package.Path)
		}
		if target.ExportName != "" {
			t.Fatalf("ExportName = %q, want empty for package summary", target.ExportName)
		}
	})

	t.Run("resolves full package export reference", func(t *testing.T) {
		target, err := Resolve(fixture.Root, "vandor/go-backend-core:usecase", packages, ModeView)
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if target.ExportName != "usecase" {
			t.Fatalf("ExportName = %q, want usecase", target.ExportName)
		}
	})

	t.Run("resolves default export for package reference in generate mode", func(t *testing.T) {
		target, err := Resolve(fixture.Root, "vandor/go-backend-core", packages, ModeGenerate)
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if target.ExportName != "default" {
			t.Fatalf("ExportName = %q, want default", target.ExportName)
		}
	})

	t.Run("resolves unique shorthand package", func(t *testing.T) {
		target, err := Resolve(fixture.Root, "api-kit", packages, ModeView)
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if target.Package.Path != "acme/api-kit" {
			t.Fatalf("Package.Path = %q, want acme/api-kit", target.Package.Path)
		}
	})

	t.Run("resolves unique shorthand export", func(t *testing.T) {
		target, err := Resolve(fixture.Root, "handler", packages, ModeView)
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if target.Package.Path != "acme/api-kit" || target.ExportName != "handler" {
			t.Fatalf("Resolve returned %+v", target)
		}
	})

	t.Run("fails on ambiguous shorthand export", func(t *testing.T) {
		ambiguousPackages := append(packages, vpkg.Package{
			Path: "vandor/web-kit",
			Manifest: vpkg.Manifest{
				Name: "vandor/web-kit",
				Exports: map[string]vpkg.Export{
					"default": {Kind: "template"},
					"handler": {Kind: "template"},
				},
			},
		})

		_, err := Resolve(fixture.Root, "handler", ambiguousPackages, ModeView)
		if err == nil {
			t.Fatal("expected Resolve to fail for ambiguous shorthand")
		}
		for _, candidate := range []string{"acme/api-kit:handler", "vandor/web-kit:handler"} {
			if !strings.Contains(err.Error(), candidate) {
				t.Fatalf("Resolve error = %v, want candidate %q", err, candidate)
			}
		}
	})
}
