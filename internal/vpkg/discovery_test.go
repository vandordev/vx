package vpkg

import (
	"path/filepath"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestDiscovery(t *testing.T) {
	t.Run("discovers packages from project vpkg directory", func(t *testing.T) {
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
			filepath.Join("vpkg", "vandor", "go-backend-core", "templates", "usecase.vxt"): "name: usecase\n",
		})

		packages, err := Discover(fixture.Root)
		if err != nil {
			t.Fatalf("Discover returned error: %v", err)
		}
		if len(packages) != 1 {
			t.Fatalf("Discover returned %d packages, want 1", len(packages))
		}

		pkg := packages[0]
		if pkg.Path != "vandor/go-backend-core" {
			t.Fatalf("package path = %q, want %q", pkg.Path, "vandor/go-backend-core")
		}
		if pkg.Manifest.Name != "vandor/go-backend-core" {
			t.Fatalf("package manifest name = %q", pkg.Manifest.Name)
		}
	})

	t.Run("ignores folders with vxt files but no manifest", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots: []string{"."},
		})
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("vpkg", "scratch", "templates", "standalone.vxt"): "name: standalone\n",
		})

		packages, err := Discover(fixture.Root)
		if err != nil {
			t.Fatalf("Discover returned error: %v", err)
		}
		if len(packages) != 0 {
			t.Fatalf("Discover returned %d packages, want 0", len(packages))
		}
	})

	t.Run("fails on invalid manifest", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots: []string{"."},
		})
		fixture.WriteFiles(t, map[string]string{
			filepath.Join("vpkg", "vandor", "broken-pack", "vpkg.yaml"): `
apiVersion: vandor.dev/v1alpha1
name: vandor/broken-pack
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: template
    templates:
      - path: templates/broken.txt
`,
			filepath.Join("vpkg", "vandor", "broken-pack", "templates", "broken.txt"): "bad\n",
		})

		_, err := Discover(fixture.Root)
		if err == nil {
			t.Fatal("expected Discover to fail for invalid manifest")
		}
	})
}
