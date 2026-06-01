package view

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vandordev/vx/internal/project"
	"github.com/vandordev/vx/internal/resolve"
	"github.com/vandordev/vx/internal/vpkg"
	"github.com/vandordev/vxt/runtime"
	"github.com/vandordev/vxt/source"
)

type Request struct {
	ProjectRoot    string
	ProjectContext project.Context
	Target         resolve.ResolvedTarget
	Input          map[string]any
	Plan           bool
}

type Result struct {
	ProjectRoot    string
	SourceClass    resolve.SourceClass
	Package        PackageSummary
	Export         ExportSummary
	RequiredInputs []InputField
	PlannedDirs    []string
	PlannedFiles   []PlannedFile
}

type PackageSummary struct {
	Path    string
	Name    string
	Version string
	Kind    string
	Exports []ExportRef
}

type ExportRef struct {
	Name string
	Kind string
}

type ExportSummary struct {
	Name          string
	Kind          string
	TemplatePaths []string
	Files         []string
	Directories   []string
	Steps         []string
	Note          string
}

type InputField struct {
	Name     string
	TypeName string
}

type PlannedFile struct {
	Path string
	Mode string
}

func Inspect(req Request) (Result, error) {
	result := Result{
		ProjectRoot: req.ProjectRoot,
		SourceClass: req.Target.SourceClass,
	}
	input := project.InjectContext(req.Input, req.ProjectContext)

	switch req.Target.SourceClass {
	case resolve.SourceClassDirectVXT:
		doc, err := compileDocument(req.Target.TemplatePath)
		if err != nil {
			return Result{}, err
		}
		result.RequiredInputs = requiredInputs(doc)
		if req.Plan {
			dirs, files, err := planDocument(doc, input)
			if err != nil {
				return Result{}, err
			}
			result.PlannedDirs = dirs
			result.PlannedFiles = files
		}
		return result, nil

	case resolve.SourceClassPackage:
		result.Package = packageSummary(req.Target.Package)
		if req.Target.ExportName == "" {
			return result, nil
		}

		export := req.Target.Package.Manifest.Exports[req.Target.ExportName]
		result.Export = exportSummary(req.Target.ExportName, export)
		if export.Kind != "template" {
			return result, nil
		}

		templatePaths := manifestTemplatePaths(req.Target.Package.Root, export)
		if len(templatePaths) == 0 {
			return Result{}, fmt.Errorf("export %q has no templates", req.Target.ExportName)
		}

		doc, err := compileDocument(templatePaths[0])
		if err != nil {
			return Result{}, err
		}
		result.RequiredInputs = requiredInputs(doc)
		if req.Plan {
			dirs, files, err := planTemplateExport(templatePaths, input)
			if err != nil {
				return Result{}, err
			}
			result.PlannedDirs = dirs
			result.PlannedFiles = files
		}
		return result, nil
	default:
		return Result{}, fmt.Errorf("unsupported source class %q", req.Target.SourceClass)
	}
}

func packageSummary(pkg vpkg.Package) PackageSummary {
	summary := PackageSummary{
		Path:    pkg.Path,
		Name:    pkg.Manifest.Name,
		Version: pkg.Manifest.Version,
		Kind:    pkg.Manifest.Kind,
	}
	for name, export := range pkg.Manifest.Exports {
		summary.Exports = append(summary.Exports, ExportRef{Name: name, Kind: export.Kind})
	}
	sort.Slice(summary.Exports, func(i, j int) bool {
		return summary.Exports[i].Name < summary.Exports[j].Name
	})
	return summary
}

func exportSummary(name string, export vpkg.Export) ExportSummary {
	summary := ExportSummary{
		Name: name,
		Kind: export.Kind,
	}
	for _, template := range export.Templates {
		summary.TemplatePaths = append(summary.TemplatePaths, template.Path)
	}
	for _, file := range export.Files {
		summary.Files = append(summary.Files, file.Path)
	}
	for _, dir := range export.Directories {
		summary.Directories = append(summary.Directories, dir.Path)
	}
	for _, step := range export.Steps {
		summary.Steps = append(summary.Steps, formatStep(step))
	}
	if export.Kind == "task" {
		summary.Note = "view-only in v0.1"
	}
	return summary
}

func manifestTemplatePaths(packageRoot string, export vpkg.Export) []string {
	paths := make([]string, 0, len(export.Templates))
	for _, template := range export.Templates {
		paths = append(paths, filepath.Join(packageRoot, template.Path))
	}
	return paths
}

func requiredInputs(doc *runtime.CompiledDocument) []InputField {
	fields := make([]InputField, 0, len(doc.Inputs))
	for _, input := range doc.Inputs {
		fields = append(fields, InputField{Name: input.Name, TypeName: input.TypeName})
	}
	return fields
}

func compileDocument(path string) (*runtime.CompiledDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %q: %w", path, err)
	}

	src := source.Source{
		ID:   path,
		Path: path,
		Text: string(data),
	}

	result := runtime.CompileDocumentWithResolver(src, fileResolver{baseDir: filepath.Dir(path)})
	if len(result.Diagnostics) > 0 {
		return nil, fmt.Errorf("%s", result.Diagnostics[0].Message)
	}
	return result.Document, nil
}

func planTemplateExport(templatePaths []string, input map[string]any) ([]string, []PlannedFile, error) {
	var plannedDirs []string
	var plannedFiles []PlannedFile
	for _, templatePath := range templatePaths {
		doc, err := compileDocument(templatePath)
		if err != nil {
			return nil, nil, err
		}
		dirs, files, err := planDocument(doc, input)
		if err != nil {
			return nil, nil, err
		}
		plannedDirs = append(plannedDirs, dirs...)
		plannedFiles = append(plannedFiles, files...)
	}
	return plannedDirs, plannedFiles, nil
}

func planDocument(doc *runtime.CompiledDocument, input map[string]any) ([]string, []PlannedFile, error) {
	validated := runtime.ValidateDocument(doc, input)
	if len(validated.Diagnostics) > 0 {
		return nil, nil, fmt.Errorf("%s", validated.Diagnostics[0].Message)
	}

	planned := runtime.PlanDocument(validated)
	if len(planned.Diagnostics) > 0 {
		return nil, nil, fmt.Errorf("%s", planned.Diagnostics[0].Message)
	}

	dirs := make([]string, 0, len(planned.Plan.Dirs))
	for _, dir := range planned.Plan.Dirs {
		dirs = append(dirs, dir.Path)
	}

	files := make([]PlannedFile, 0, len(planned.Plan.Files))
	for _, file := range planned.Plan.Files {
		files = append(files, PlannedFile{
			Path: file.Path,
			Mode: file.Mode,
		})
	}

	return dirs, files, nil
}

func formatStep(step map[string]any) string {
	parts := make([]string, 0, len(step))
	if stepType, ok := step["type"]; ok {
		parts = append(parts, fmt.Sprintf("type=%v", stepType))
	}
	keys := make([]string, 0, len(step))
	for key := range step {
		if key == "type" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, step[key]))
	}
	return strings.Join(parts, " ")
}

type fileResolver struct {
	baseDir string
}

func (r fileResolver) Resolve(path string) (source.Source, error) {
	resolvedPath := filepath.Join(r.baseDir, path)
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return source.Source{}, err
	}
	return source.Source{
		ID:   resolvedPath,
		Path: resolvedPath,
		Text: string(data),
	}, nil
}
