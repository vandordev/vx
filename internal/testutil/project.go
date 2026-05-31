package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

type ProjectLayout struct {
	VPKGRoots   []string
	Directories []string
}

type ProjectFixture struct {
	Root string
}

func NewProjectFixture(t *testing.T, layout ProjectLayout) ProjectFixture {
	t.Helper()

	root := t.TempDir()
	for _, relativeRoot := range layout.VPKGRoots {
		path := filepath.Join(root, relativeRoot, "vpkg")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("mkdir vpkg root %s: %v", path, err)
		}
	}

	for _, relativeDir := range layout.Directories {
		path := filepath.Join(root, relativeDir)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("mkdir project directory %s: %v", path, err)
		}
	}

	return ProjectFixture{Root: root}
}

func (f ProjectFixture) Path(parts ...string) string {
	segments := append([]string{f.Root}, parts...)
	return filepath.Join(segments...)
}
