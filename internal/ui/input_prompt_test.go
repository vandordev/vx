package ui

import "testing"

func TestInputPromptLabelsIncludeNameAndType(t *testing.T) {
	fields := []PromptField{
		{Name: "name", TypeName: "string"},
		{Name: "enabled", TypeName: "bool"},
	}

	labels := InputPromptLabels(fields)

	if len(labels) != 2 {
		t.Fatalf("labels = %#v", labels)
	}
	if labels[0] != "name (string)" {
		t.Fatalf("labels[0] = %q", labels[0])
	}
	if labels[1] != "enabled (bool)" {
		t.Fatalf("labels[1] = %q", labels[1])
	}
}
