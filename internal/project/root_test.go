package project

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestFindRoot(t *testing.T) {
	t.Run("returns current directory when it contains vpkg", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots: []string{"."},
		})

		root, err := FindRoot(fixture.Root)
		if err != nil {
			t.Fatalf("FindRoot returned error: %v", err)
		}
		if root != fixture.Root {
			t.Fatalf("FindRoot = %q, want %q", root, fixture.Root)
		}
	})

	t.Run("returns nearest parent containing vpkg", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			VPKGRoots:   []string{".", filepath.Join("apps", "api")},
			Directories: []string{filepath.Join("apps", "api", "internal", "handlers")},
		})

		start := fixture.Path("apps", "api", "internal", "handlers")
		root, err := FindRoot(start)
		if err != nil {
			t.Fatalf("FindRoot returned error: %v", err)
		}

		want := fixture.Path("apps", "api")
		if root != want {
			t.Fatalf("FindRoot = %q, want %q", root, want)
		}
	})

	t.Run("returns typed error when no parent contains vpkg", func(t *testing.T) {
		fixture := testutil.NewProjectFixture(t, testutil.ProjectLayout{
			Directories: []string{filepath.Join("pkg", "feature")},
		})

		_, err := FindRoot(fixture.Path("pkg", "feature"))
		if !errors.Is(err, ErrRootNotFound) {
			t.Fatalf("expected ErrRootNotFound, got %v", err)
		}
	})
}
