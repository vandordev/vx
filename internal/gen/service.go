package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/vandordev/vx/internal/resolve"
	"github.com/vandordev/vx/internal/vpkg"
	"github.com/vandordev/vxt/runtime"
	"github.com/vandordev/vxt/source"
	"github.com/vandordev/vxt/write"
)

type Request struct {
	ProjectRoot    string
	Target         resolve.ResolvedTarget
	Input          map[string]any
	Apply          bool
	NonInteractive bool
}

type Result struct {
	ProjectRoot      string
	SourceClass      resolve.SourceClass
	ExportName       string
	PlannedDirs      []string
	PlannedFiles     []PlannedFile
	ConflictingFiles []string
	Applied          bool
}

type PlannedFile struct {
	Path string
	Mode string
}

func Generate(req Request) (Result, error) {
	result := Result{
		ProjectRoot: req.ProjectRoot,
		SourceClass: req.Target.SourceClass,
		ExportName:  req.Target.ExportName,
	}

	plan, err := buildPlan(req.Target, req.Input)
	if err != nil {
		return Result{}, err
	}

	for _, dir := range plan.Dirs {
		result.PlannedDirs = append(result.PlannedDirs, dir.Path)
	}
	for _, file := range plan.Files {
		result.PlannedFiles = append(result.PlannedFiles, PlannedFile{
			Path: file.Path,
			Mode: file.Mode,
		})
		if pathExists(filepath.Join(req.ProjectRoot, file.Path)) {
			result.ConflictingFiles = append(result.ConflictingFiles, file.Path)
		}
	}
	sort.Strings(result.ConflictingFiles)

	if !req.Apply {
		return result, nil
	}
	if len(result.ConflictingFiles) > 0 {
		return Result{}, fmt.Errorf("planned output already exists: %s", result.ConflictingFiles[0])
	}

	target := write.NewFilesystemTarget(req.ProjectRoot)
	if _, err := runtime.WritePlan(plan, target); err != nil {
		return Result{}, err
	}

	result.Applied = true
	return result, nil
}

func buildPlan(target resolve.ResolvedTarget, input map[string]any) (runtime.Plan, error) {
	switch target.SourceClass {
	case resolve.SourceClassDirectVXT:
		return planTemplates([]string{target.TemplatePath}, input)
	case resolve.SourceClassPackage:
		if target.Export.Kind != "template" {
			return runtime.Plan{}, fmt.Errorf("%s exports are view-only in v0.1", target.Export.Kind)
		}
		return planTemplates(templatePaths(target.Package.Root, target.Export), input)
	default:
		return runtime.Plan{}, fmt.Errorf("unsupported source class %q", target.SourceClass)
	}
}

func templatePaths(packageRoot string, export vpkg.Export) []string {
	paths := make([]string, 0, len(export.Templates))
	for _, template := range export.Templates {
		paths = append(paths, filepath.Join(packageRoot, template.Path))
	}
	return paths
}

func planTemplates(templatePaths []string, input map[string]any) (runtime.Plan, error) {
	combined := runtime.Plan{}
	for _, templatePath := range templatePaths {
		doc, err := compileDocument(templatePath)
		if err != nil {
			return runtime.Plan{}, err
		}

		validated := runtime.ValidateDocument(doc, input)
		if len(validated.Diagnostics) > 0 {
			return runtime.Plan{}, fmt.Errorf("%s", validated.Diagnostics[0].Message)
		}

		planned := runtime.PlanDocument(validated)
		if len(planned.Diagnostics) > 0 {
			return runtime.Plan{}, fmt.Errorf("%s", planned.Diagnostics[0].Message)
		}

		combined.Dirs = append(combined.Dirs, planned.Plan.Dirs...)
		combined.Files = append(combined.Files, planned.Plan.Files...)
		combined.PlannedHooks = append(combined.PlannedHooks, planned.Plan.PlannedHooks...)
	}
	return combined, nil
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

func pathExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
