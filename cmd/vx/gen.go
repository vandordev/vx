package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	gensvc "github.com/vandordev/vx/internal/gen"
	"github.com/vandordev/vx/internal/input"
	"github.com/vandordev/vx/internal/project"
	"github.com/vandordev/vx/internal/resolve"
	"github.com/vandordev/vx/internal/vpkg"
)

type genOptions struct {
	apply          bool
	valuesPath     string
	sets           []string
	asJSON         bool
	nonInteractive bool
}

func newGenCmd() *cobra.Command {
	opts := &genOptions{}
	cmd := &cobra.Command{
		Use:     "gen <target>",
		Aliases: []string{"generate"},
		Short:   "Preview or apply generation for a local vpkg export or .vxt file",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGen(cmd, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.apply, "apply", false, "write the planned output to the project root")
	cmd.Flags().StringVar(&opts.valuesPath, "values", "", "path to a YAML or JSON values file")
	cmd.Flags().StringArrayVar(&opts.sets, "set", nil, "set one input value with key=value syntax")
	cmd.Flags().BoolVar(&opts.asJSON, "json", false, "print machine-readable JSON output")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "fail instead of prompting for missing input")
	return cmd
}

func runGen(cmd *cobra.Command, opts *genOptions, targetArg string) error {
	projectRoot, err := project.FindRoot(".")
	if err != nil {
		return err
	}

	packages, err := vpkg.Discover(projectRoot)
	if err != nil {
		return err
	}

	target, err := resolve.Resolve(projectRoot, targetArg, packages, resolve.ModeGenerate)
	if err != nil {
		return err
	}

	values, err := input.Load(opts.valuesPath, opts.sets)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectContext, err := project.DetectContext(projectRoot, cwd)
	if err != nil {
		return err
	}

	result, err := gensvc.Generate(gensvc.Request{
		ProjectRoot:    projectRoot,
		ProjectContext: projectContext,
		Target:         target,
		Input:          values,
		Apply:          opts.apply,
		NonInteractive: opts.nonInteractive,
	})
	if err != nil {
		return err
	}

	if opts.asJSON {
		return renderGenJSON(cmd, result)
	}

	renderGenText(cmd, targetArg, result)
	return nil
}

func renderGenText(cmd *cobra.Command, targetArg string, result gensvc.Result) {
	var lines []string
	lines = append(lines, fmt.Sprintf("Preview: %s", targetArg))
	lines = append(lines, fmt.Sprintf("Project Root: %s", result.ProjectRoot))

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

	if len(result.ConflictingFiles) > 0 {
		lines = append(lines, "Conflicts:")
		for _, path := range result.ConflictingFiles {
			lines = append(lines, fmt.Sprintf("  - %s", path))
		}
	}

	lines = append(lines, fmt.Sprintf("Applied: %t", result.Applied))
	cmd.Println(strings.Join(lines, "\n"))
}

func renderGenJSON(cmd *cobra.Command, result gensvc.Result) error {
	payload := struct {
		ProjectRoot      string               `json:"projectRoot"`
		SourceClass      resolve.SourceClass  `json:"sourceClass"`
		ExportName       string               `json:"exportName,omitempty"`
		PlannedDirs      []string             `json:"plannedDirs,omitempty"`
		PlannedFiles     []gensvc.PlannedFile `json:"plannedFiles,omitempty"`
		ConflictingFiles []string             `json:"conflictingFiles,omitempty"`
		Applied          bool                 `json:"applied"`
	}{
		ProjectRoot:      result.ProjectRoot,
		SourceClass:      result.SourceClass,
		ExportName:       result.ExportName,
		PlannedDirs:      result.PlannedDirs,
		PlannedFiles:     result.PlannedFiles,
		ConflictingFiles: result.ConflictingFiles,
		Applied:          result.Applied,
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}
