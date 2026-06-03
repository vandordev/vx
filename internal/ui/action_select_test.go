package ui

import "testing"

func TestGenerationActionOptions(t *testing.T) {
	options := GenerationActionOptions()

	if len(options) != 2 {
		t.Fatalf("len(options) = %d, want 2", len(options))
	}
	if options[0].Value != GenerationActionPreview {
		t.Fatalf("first option = %q, want preview", options[0].Value)
	}
	if options[1].Value != GenerationActionApply {
		t.Fatalf("second option = %q, want apply", options[1].Value)
	}
}

func TestDefaultGenerationAction(t *testing.T) {
	if got := DefaultGenerationAction(); got != GenerationActionPreview {
		t.Fatalf("DefaultGenerationAction() = %q, want %q", got, GenerationActionPreview)
	}
}
