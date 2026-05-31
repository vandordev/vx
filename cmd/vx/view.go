package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vandordev/vx/internal/input"
	"github.com/vandordev/vx/internal/project"
	"github.com/vandordev/vx/internal/resolve"
	viewsvc "github.com/vandordev/vx/internal/view"
	"github.com/vandordev/vx/internal/vpkg"
)

type viewOptions struct {
	plan           bool
	valuesPath     string
	sets           []string
	asJSON         bool
	nonInteractive bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}
	cmd := &cobra.Command{
		Use:   "view <target>",
		Short: "Inspect a local vpkg package, export, or direct .vxt file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.plan, "plan", false, "compile, validate, and plan template output")
	cmd.Flags().StringVar(&opts.valuesPath, "values", "", "path to a YAML or JSON values file")
	cmd.Flags().StringArrayVar(&opts.sets, "set", nil, "set one input value with key=value syntax")
	cmd.Flags().BoolVar(&opts.asJSON, "json", false, "print machine-readable JSON output")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "fail instead of prompting for missing input")
	return cmd
}

func runView(cmd *cobra.Command, opts *viewOptions, targetArg string) error {
	projectRoot, err := project.FindRoot(".")
	if err != nil {
		return err
	}

	packages, err := vpkg.Discover(projectRoot)
	if err != nil {
		return err
	}

	target, err := resolve.Resolve(projectRoot, targetArg, packages, resolve.ModeView)
	if err != nil {
		return err
	}

	values, err := input.Load(opts.valuesPath, opts.sets)
	if err != nil {
		return err
	}

	result, err := viewsvc.Inspect(viewsvc.Request{
		ProjectRoot: projectRoot,
		Target:      target,
		Input:       values,
		Plan:        opts.plan,
	})
	if err != nil {
		return err
	}

	if opts.asJSON {
		return renderViewJSON(cmd, result)
	}

	renderViewText(cmd, result)
	return nil
}

func renderViewText(cmd *cobra.Command, result viewsvc.Result) {
	var lines []string
	if result.Package.Path != "" {
		lines = append(lines, fmt.Sprintf("Package: %s", result.Package.Name))
		lines = append(lines, fmt.Sprintf("Version: %s", result.Package.Version))
		lines = append(lines, fmt.Sprintf("Kind: %s", result.Package.Kind))
	}

	if result.Export.Name == "" {
		if len(result.Package.Exports) > 0 {
			lines = append(lines, "Exports:")
			for _, export := range result.Package.Exports {
				lines = append(lines, fmt.Sprintf("  - %s (%s)", export.Name, export.Kind))
			}
		}
		cmd.Println(strings.Join(lines, "\n"))
		return
	}

	lines = append(lines, fmt.Sprintf("Export: %s", result.Export.Name))
	lines = append(lines, fmt.Sprintf("Kind: %s", result.Export.Kind))

	if len(result.Export.TemplatePaths) > 0 {
		lines = append(lines, "Templates:")
		for _, path := range result.Export.TemplatePaths {
			lines = append(lines, fmt.Sprintf("  - %s", path))
		}
	}

	if len(result.RequiredInputs) > 0 {
		lines = append(lines, "Required Inputs:")
		for _, input := range result.RequiredInputs {
			lines = append(lines, fmt.Sprintf("  - %s (%s)", input.Name, input.TypeName))
		}
	}

	if len(result.Export.Files) > 0 {
		lines = append(lines, "Files:")
		for _, path := range result.Export.Files {
			lines = append(lines, fmt.Sprintf("  - %s", path))
		}
	}

	if len(result.Export.Directories) > 0 {
		lines = append(lines, "Directories:")
		for _, path := range result.Export.Directories {
			lines = append(lines, fmt.Sprintf("  - %s", path))
		}
	}

	if len(result.Export.Steps) > 0 {
		lines = append(lines, "Steps:")
		for idx, step := range result.Export.Steps {
			lines = append(lines, fmt.Sprintf("  %d. %s", idx+1, step))
		}
	}

	if result.Export.Note != "" {
		lines = append(lines, fmt.Sprintf("Note: %s", result.Export.Note))
	}

	if len(result.PlannedDirs) > 0 {
		lines = append(lines, "Planned Directories:")
		for _, path := range result.PlannedDirs {
			lines = append(lines, fmt.Sprintf("  - %s", path))
		}
	}

	if len(result.PlannedFiles) > 0 {
		lines = append(lines, "Planned Files:")
		for _, file := range result.PlannedFiles {
			lines = append(lines, fmt.Sprintf("  - %s (%s)", file.Path, file.Mode))
		}
	}

	cmd.Println(strings.Join(lines, "\n"))
}

func renderViewJSON(cmd *cobra.Command, result viewsvc.Result) error {
	payload := struct {
		ProjectRoot    string                 `json:"projectRoot"`
		SourceClass    resolve.SourceClass    `json:"sourceClass"`
		Package        viewsvc.PackageSummary `json:"package"`
		Export         viewsvc.ExportSummary  `json:"export"`
		RequiredInputs []viewsvc.InputField   `json:"requiredInputs"`
		PlannedDirs    []string               `json:"plannedDirs,omitempty"`
		PlannedFiles   []viewsvc.PlannedFile  `json:"plannedFiles,omitempty"`
	}{
		ProjectRoot:    result.ProjectRoot,
		SourceClass:    result.SourceClass,
		Package:        result.Package,
		Export:         result.Export,
		RequiredInputs: result.RequiredInputs,
		PlannedDirs:    result.PlannedDirs,
		PlannedFiles:   result.PlannedFiles,
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}
