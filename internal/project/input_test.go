package project

import "testing"

func TestInjectContextPreservesUserInputAndReplacesProject(t *testing.T) {
	input := map[string]any{
		"name": "billing",
		"project": map[string]any{
			"root": "user-supplied",
		},
	}

	merged := InjectContext(input, Context{Root: "/repo"})

	if merged["name"] != "billing" {
		t.Fatalf("name = %#v", merged["name"])
	}
	if inputProject := input["project"].(map[string]any); inputProject["root"] != "user-supplied" {
		t.Fatalf("input project root mutated to %#v", inputProject["root"])
	}
	mergedProject, ok := merged["project"].(map[string]any)
	if !ok {
		t.Fatalf("merged project = %#v", merged["project"])
	}
	if mergedProject["root"] != "/repo" {
		t.Fatalf("merged project root = %#v", mergedProject["root"])
	}
}

func TestInjectContextHandlesNilInput(t *testing.T) {
	merged := InjectContext(nil, Context{Root: "/repo"})

	projectValues, ok := merged["project"].(map[string]any)
	if !ok {
		t.Fatalf("project values = %#v", merged["project"])
	}
	if projectValues["root"] != "/repo" {
		t.Fatalf("project.root = %#v", projectValues["root"])
	}
}
