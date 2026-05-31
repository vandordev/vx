package main

import (
	"strings"
	"testing"

	"github.com/vandordev/vx/internal/domain"
)

func TestRenderConfigTemplateOmitsInteractiveDefault(t *testing.T) {
	content := renderConfigTemplate(domain.DefaultConfig())

	if strings.Contains(content, "interactive_default") {
		t.Fatalf("expected config template to omit interactive_default, content:\n%s", content)
	}

	if strings.Contains(strings.ToLower(content), "interactive mode") {
		t.Fatalf("expected config template to omit interactive mode wording, content:\n%s", content)
	}
}
