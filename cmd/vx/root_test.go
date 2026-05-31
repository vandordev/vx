package main

import (
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestRootCommandHasVersion(t *testing.T) {
	cmd := newRootCmd()
	versionFlag := cmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("expected --version flag to be registered")
	}
	if versionFlag.Shorthand != "v" {
		t.Fatalf("expected shorthand -v, got %q", versionFlag.Shorthand)
	}
}

func TestRootCommandHasConfig(t *testing.T) {
	cmd := newRootCmd()
	configFlag := cmd.Flags().Lookup("config")
	if configFlag == nil {
		t.Fatal("expected --config flag to be registered")
	}
}

func TestResolvedVersion(t *testing.T) {
	ver := resolvedVersion()
	if ver == "" {
		t.Fatal("expected version to be non-empty")
	}
}

func TestRootCommandWithoutArgsPrintsOverview(t *testing.T) {
	cmd := newRootCmd()

	output, err := testutil.RunCLI(t, cmd)
	if err != nil {
		t.Fatalf("expected root command without args to succeed, got error: %v\noutput:\n%s", err, output)
	}

	for _, snippet := range []string{
		"vx view <target>",
		"vx gen <target>",
	} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected root overview to contain %q, output:\n%s", snippet, output)
		}
	}

	legacyTitle := "Directory" + ":"
	if strings.Contains(output, legacyTitle) {
		t.Fatalf("expected root overview to avoid the legacy title, output:\n%s", output)
	}
}
