package vpkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestValidation(t *testing.T) {
	t.Run("loads a valid manifest", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
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
			filepath.Join("templates", "usecase.vxt"): "name: usecase\n",
		})

		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			t.Fatalf("LoadManifest returned error: %v", err)
		}
		if manifest.Name != "vandor/go-backend-core" {
			t.Fatalf("LoadManifest name = %q", manifest.Name)
		}
		if manifest.Exports["default"].Kind != "template" {
			t.Fatalf("LoadManifest export kind = %q", manifest.Exports["default"].Kind)
		}
	})

	t.Run("requires package fields", func(t *testing.T) {
		cases := []struct {
			name    string
			content string
			want    string
		}{
			{
				name: "apiVersion",
				content: `
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports: {}
`,
				want: "apiVersion",
			},
			{
				name: "name",
				content: `
apiVersion: vandor.dev/v1alpha1
version: 0.1.0
kind: template-pack
exports: {}
`,
				want: "name",
			},
			{
				name: "version",
				content: `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
kind: template-pack
exports: {}
`,
				want: "version",
			},
			{
				name: "kind",
				content: `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
exports: {}
`,
				want: "kind",
			},
			{
				name: "exports",
				content: `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
`,
				want: "exports",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
					"vpkg.yaml": tc.content,
				})

				_, err := LoadManifest(manifestPath)
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("LoadManifest error = %v, want mention of %q", err, tc.want)
				}
			})
		}
	})

	t.Run("rejects invalid package kind", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: unsupported
exports:
  default:
    kind: preset
`,
		})

		_, err := LoadManifest(manifestPath)
		if err == nil || !strings.Contains(err.Error(), "kind") {
			t.Fatalf("LoadManifest error = %v, want invalid package kind", err)
		}
	})

	t.Run("rejects invalid export kind", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: unsupported
`,
		})

		_, err := LoadManifest(manifestPath)
		if err == nil || !strings.Contains(err.Error(), "export") {
			t.Fatalf("LoadManifest error = %v, want invalid export kind", err)
		}
	})

	t.Run("rejects package name mismatch with path", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
apiVersion: vandor.dev/v1alpha1
name: vandor/other-package
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: preset
`,
		})

		_, err := LoadManifest(manifestPath)
		if err == nil || !strings.Contains(err.Error(), "path") {
			t.Fatalf("LoadManifest error = %v, want path mismatch", err)
		}
	})

	t.Run("rejects template export without templates", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: template
`,
		})

		_, err := LoadManifest(manifestPath)
		if err == nil || !strings.Contains(err.Error(), "templates") {
			t.Fatalf("LoadManifest error = %v, want templates validation error", err)
		}
	})

	t.Run("rejects non vxt template paths", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: template
    templates:
      - path: templates/usecase.txt
`,
			filepath.Join("templates", "usecase.txt"): "not a vxt file\n",
		})

		_, err := LoadManifest(manifestPath)
		if err == nil || !strings.Contains(err.Error(), ".vxt") {
			t.Fatalf("LoadManifest error = %v, want .vxt validation error", err)
		}
	})

	t.Run("rejects template paths escaping package root", func(t *testing.T) {
		manifestPath := writePackageFixture(t, filepath.Join("vpkg", "vandor", "go-backend-core"), map[string]string{
			"vpkg.yaml": `
apiVersion: vandor.dev/v1alpha1
name: vandor/go-backend-core
version: 0.1.0
kind: template-pack
exports:
  default:
    kind: template
    templates:
      - path: ../shared/usecase.vxt
`,
		})

		_, err := LoadManifest(manifestPath)
		if err == nil || !strings.Contains(err.Error(), "package root") {
			t.Fatalf("LoadManifest error = %v, want package root validation error", err)
		}
	})
}

func writePackageFixture(t *testing.T, packageDir string, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	absPackageDir := filepath.Join(root, packageDir)
	for relativePath, content := range files {
		absPath := filepath.Join(absPackageDir, relativePath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(absPath), err)
		}
		if err := os.WriteFile(absPath, []byte(strings.TrimLeft(content, "\n")), 0o644); err != nil {
			t.Fatalf("write %s: %v", absPath, err)
		}
	}

	return filepath.Join(absPackageDir, "vpkg.yaml")
}
