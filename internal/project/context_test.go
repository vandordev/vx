package project

import (
	"path/filepath"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestDetectContextAlwaysIncludesProjectRoot(t *testing.T) {
	fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
		VPKGRoots: []string{"."},
	})

	ctx, err := DetectContext(fixture.Root, fixture.Root)
	if err != nil {
		t.Fatalf("DetectContext returned error: %v", err)
	}

	if ctx.Root != fixture.Root {
		t.Fatalf("Root = %q, want %q", ctx.Root, fixture.Root)
	}
	if ctx.Language != "" {
		t.Fatalf("Language = %q, want empty when no language is detected", ctx.Language)
	}
	if ctx.Go != nil {
		t.Fatalf("Go = %#v, want nil when no Go module is detected", ctx.Go)
	}
}

func TestDetectContextUsesNearestInRootGoMod(t *testing.T) {
	fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
		VPKGRoots: []string{"."},
		Directories: []string{
			filepath.Join("services", "billing", "internal", "app"),
		},
	})
	fixture.WriteFiles(t, map[string]string{
		"go.mod": "module github.com/acme/platform\n\ngo 1.25\n",
		filepath.Join("services", "billing", "go.mod"): "module github.com/acme/platform/services/billing\n\ngo 1.25\n",
	})

	ctx, err := DetectContext(fixture.Root, fixture.Path("services", "billing", "internal", "app"))
	if err != nil {
		t.Fatalf("DetectContext returned error: %v", err)
	}

	if ctx.Language != "go" {
		t.Fatalf("Language = %q, want go", ctx.Language)
	}
	if ctx.Go == nil {
		t.Fatal("Go context is nil, want detected module")
	}
	if ctx.Go.Module != "github.com/acme/platform/services/billing" {
		t.Fatalf("Go.Module = %q", ctx.Go.Module)
	}
	if ctx.Go.ModuleRoot != filepath.Join("services", "billing") {
		t.Fatalf("Go.ModuleRoot = %q", ctx.Go.ModuleRoot)
	}
}

func TestDetectContextOmitsGoContextWhenNoInRootModuleExists(t *testing.T) {
	t.Run("without go.mod", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots: []string{"."},
			Directories: []string{
				filepath.Join("services", "billing"),
			},
		})

		ctx, err := DetectContext(fixture.Root, fixture.Path("services", "billing"))
		if err != nil {
			t.Fatalf("DetectContext returned error: %v", err)
		}

		assertNoGoContext(t, ctx)
	})

	t.Run("ignores go.mod above project root", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots: []string{filepath.Join("repo")},
			Directories: []string{
				filepath.Join("repo", "services", "billing"),
			},
		})
		fixture.WriteFiles(t, map[string]string{
			"go.mod": "module github.com/acme/outside\n\ngo 1.25\n",
		})

		ctx, err := DetectContext(fixture.Path("repo"), fixture.Path("repo", "services", "billing"))
		if err != nil {
			t.Fatalf("DetectContext returned error: %v", err)
		}

		assertNoGoContext(t, ctx)
	})

	t.Run("ignores go.mod without module directive", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots: []string{"."},
			Directories: []string{
				filepath.Join("services", "billing"),
			},
		})
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("services", "billing", "go.mod"): "go 1.25\n",
		})

		ctx, err := DetectContext(fixture.Root, fixture.Path("services", "billing"))
		if err != nil {
			t.Fatalf("DetectContext returned error: %v", err)
		}

		assertNoGoContext(t, ctx)
	})
}

func TestContextValuesIncludesSparseProjectContext(t *testing.T) {
	ctx := Context{
		Root:     "/repo",
		Language: "go",
		Go: &GoContext{
			Module:     "github.com/acme/platform/services/billing",
			ModuleRoot: filepath.Join("services", "billing"),
		},
	}

	values := ctx.Values()
	projectValues, ok := values["project"].(map[string]any)
	if !ok {
		t.Fatalf("project values = %#v", values["project"])
	}
	goValues, ok := projectValues["go"].(map[string]any)
	if !ok {
		t.Fatalf("go values = %#v", projectValues["go"])
	}

	if projectValues["root"] != "/repo" {
		t.Fatalf("project.root = %#v", projectValues["root"])
	}
	if projectValues["language"] != "go" {
		t.Fatalf("project.language = %#v", projectValues["language"])
	}
	if goValues["module"] != "github.com/acme/platform/services/billing" {
		t.Fatalf("project.go.module = %#v", goValues["module"])
	}
	if goValues["module_root"] != filepath.Join("services", "billing") {
		t.Fatalf("project.go.module_root = %#v", goValues["module_root"])
	}
}

func TestContextValuesOmitsUndetectedFields(t *testing.T) {
	values := Context{Root: "/repo"}.Values()
	projectValues, ok := values["project"].(map[string]any)
	if !ok {
		t.Fatalf("project values = %#v", values["project"])
	}

	if projectValues["root"] != "/repo" {
		t.Fatalf("project.root = %#v", projectValues["root"])
	}
	if _, ok := projectValues["language"]; ok {
		t.Fatalf("project.language present, values = %#v", projectValues)
	}
	if _, ok := projectValues["go"]; ok {
		t.Fatalf("project.go present, values = %#v", projectValues)
	}
}

func assertNoGoContext(t *testing.T, ctx Context) {
	t.Helper()

	if ctx.Language != "" {
		t.Fatalf("Language = %q, want empty", ctx.Language)
	}
	if ctx.Go != nil {
		t.Fatalf("Go = %#v, want nil", ctx.Go)
	}
}
