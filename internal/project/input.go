package project

func InjectContext(input map[string]any, ctx Context) map[string]any {
	merged := make(map[string]any, len(input)+1)
	for key, value := range input {
		if key == "project" {
			continue
		}
		merged[key] = value
	}

	merged["project"] = ctx.Values()["project"]
	return merged
}
