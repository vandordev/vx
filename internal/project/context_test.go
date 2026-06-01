package project

import (
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
