package main

import (
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
