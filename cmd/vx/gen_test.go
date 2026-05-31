package main

import (
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/testutil"
)

func TestGenCommandIsRegistered(t *testing.T) {
	cmd := newRootCmd()

	genCmd, _, err := cmd.Find([]string{"gen"})
	if err != nil {
		t.Fatalf("expected gen command to be registered, got error: %v", err)
	}
	if genCmd == nil || genCmd.Name() != "gen" {
		t.Fatalf("expected to resolve gen command, got %#v", genCmd)
	}
}

func TestGenerateAliasResolvesToGen(t *testing.T) {
	cmd := newRootCmd()

	genCmd, _, err := cmd.Find([]string{"gen"})
	if err != nil {
		t.Fatalf("expected gen command to be registered, got error: %v", err)
	}

	aliasCmd, _, err := cmd.Find([]string{"generate"})
	if err != nil {
		t.Fatalf("expected generate alias to resolve, got error: %v", err)
	}
	if aliasCmd == nil || aliasCmd.Name() != "gen" {
		t.Fatalf("expected generate alias to resolve to gen, got %#v", aliasCmd)
	}
	if aliasCmd != genCmd {
		t.Fatalf("expected generate alias to resolve to the gen command instance")
	}
}

func TestGenCommandHelp(t *testing.T) {
	cmd := newRootCmd()

	output, err := testutil.RunCLI(t, cmd, "gen", "--help")
	if err != nil {
		t.Fatalf("expected gen help to succeed, got error: %v\noutput:\n%s", err, output)
	}

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output for gen command, output:\n%s", output)
	}
}

func TestGenerateAliasHelp(t *testing.T) {
	cmd := newRootCmd()

	output, err := testutil.RunCLI(t, cmd, "generate", "--help")
	if err != nil {
		t.Fatalf("expected generate alias help to succeed, got error: %v\noutput:\n%s", err, output)
	}

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output for generate alias, output:\n%s", output)
	}
}
