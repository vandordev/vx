package main

import (
	"testing"
)

func TestCLIRuns(t *testing.T) {
	cmd := newRootCmd()
	if cmd == nil {
		t.Fatal("expected root command to be created")
	}

	if cmd.Name() != "vx" {
		t.Fatalf("expected root command name vx, got %q", cmd.Name())
	}
}
