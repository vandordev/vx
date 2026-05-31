package main

import (
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/domain"
)

func TestRenderConfigTemplateOmitsInteractiveDefault(t *testing.T) {
	content := renderConfigTemplate(domain.DefaultConfig())

	legacyKey := "interactive" + "_default"
	if strings.Contains(content, legacyKey) {
		t.Fatalf("expected config template to omit the legacy startup key, content:\n%s", content)
	}

	if strings.Contains(strings.ToLower(content), "interactive mode") {
		t.Fatalf("expected config template to omit startup-mode wording, content:\n%s", content)
	}
}
