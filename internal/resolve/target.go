package resolve

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vandordev/vx/internal/vpkg"
)

type Mode string

const (
	ModeView     Mode = "view"
	ModeGenerate Mode = "generate"
)

type SourceClass string

const (
	SourceClassPackage   SourceClass = "package"
	SourceClassDirectVXT SourceClass = "direct-vxt"
)

type ResolvedTarget struct {
	SourceClass  SourceClass
	Package      vpkg.Package
	ExportName   string
	Export       vpkg.Export
	TemplatePath string
}

func Resolve(projectRoot string, target string, packages []vpkg.Package, mode Mode) (ResolvedTarget, error) {
	if strings.HasSuffix(target, ".vxt") {
		return resolveDirectTemplate(projectRoot, target)
	}

	packageRef, exportRef := splitTarget(target)
	if resolved, ok, err := resolveExactPackage(packageRef, exportRef, packages, mode); ok || err != nil {
		return resolved, err
	}
	if resolved, ok, err := resolveShorthandPackage(target, packages, mode); ok || err != nil {
		return resolved, err
	}
	if resolved, ok, err := resolveShorthandExport(target, packages); ok || err != nil {
		return resolved, err
	}

	return ResolvedTarget{}, fmt.Errorf("target %q did not match a local vpkg package, export, or .vxt path", target)
}

func resolveDirectTemplate(projectRoot, target string) (ResolvedTarget, error) {
	absTarget := target
	if !filepath.IsAbs(target) {
		absTarget = filepath.Join(projectRoot, target)
	}

	absTarget, err := filepath.Abs(absTarget)
	if err != nil {
		return ResolvedTarget{}, fmt.Errorf("resolve direct template path: %w", err)
	}
	if !isWithinRoot(projectRoot, absTarget) {
		return ResolvedTarget{}, fmt.Errorf("direct template %q must be inside project root %q", absTarget, projectRoot)
	}
	return ResolvedTarget{
		SourceClass:  SourceClassDirectVXT,
		TemplatePath: absTarget,
	}, nil
}

func resolveExactPackage(packageRef, exportRef string, packages []vpkg.Package, mode Mode) (ResolvedTarget, bool, error) {
	for _, pkg := range packages {
		if pkg.Path != packageRef {
			continue
		}
		resolved, err := buildPackageTarget(pkg, exportRef, mode)
		return resolved, true, err
	}
	return ResolvedTarget{}, false, nil
}

func resolveShorthandPackage(target string, packages []vpkg.Package, mode Mode) (ResolvedTarget, bool, error) {
	var matches []vpkg.Package
	for _, pkg := range packages {
		if filepath.Base(pkg.Path) == target {
			matches = append(matches, pkg)
		}
	}
	if len(matches) == 0 {
		return ResolvedTarget{}, false, nil
	}
	if len(matches) > 1 {
		return ResolvedTarget{}, true, fmt.Errorf("package shorthand %q is ambiguous: %s", target, joinCandidates(packageCandidates(matches)))
	}
	resolved, err := buildPackageTarget(matches[0], "", mode)
	return resolved, true, err
}

func resolveShorthandExport(target string, packages []vpkg.Package) (ResolvedTarget, bool, error) {
	type candidate struct {
		pkg    vpkg.Package
		export string
	}
	var matches []candidate
	for _, pkg := range packages {
		if _, ok := pkg.Manifest.Exports[target]; ok {
			matches = append(matches, candidate{pkg: pkg, export: target})
		}
	}
	if len(matches) == 0 {
		return ResolvedTarget{}, false, nil
	}
	if len(matches) > 1 {
		var candidates []string
		for _, match := range matches {
			candidates = append(candidates, fmt.Sprintf("%s:%s", match.pkg.Path, match.export))
		}
		sort.Strings(candidates)
		return ResolvedTarget{}, true, fmt.Errorf("export shorthand %q is ambiguous: %s", target, joinCandidates(candidates))
	}

	export := matches[0].pkg.Manifest.Exports[target]
	return ResolvedTarget{
		SourceClass: SourceClassPackage,
		Package:     matches[0].pkg,
		ExportName:  target,
		Export:      export,
	}, true, nil
}

func buildPackageTarget(pkg vpkg.Package, exportRef string, mode Mode) (ResolvedTarget, error) {
	if exportRef == "" {
		if mode == ModeView {
			return ResolvedTarget{
				SourceClass: SourceClassPackage,
				Package:     pkg,
			}, nil
		}
		exportRef = "default"
	}

	export, ok := pkg.Manifest.Exports[exportRef]
	if !ok {
		return ResolvedTarget{}, fmt.Errorf("package %q does not export %q", pkg.Path, exportRef)
	}

	return ResolvedTarget{
		SourceClass: SourceClassPackage,
		Package:     pkg,
		ExportName:  exportRef,
		Export:      export,
	}, nil
}

func splitTarget(target string) (packageRef string, exportRef string) {
	parts := strings.SplitN(target, ":", 2)
	packageRef = parts[0]
	if len(parts) == 2 {
		exportRef = parts[1]
	}
	return packageRef, exportRef
}

func isWithinRoot(projectRoot, targetPath string) bool {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, targetPath)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func packageCandidates(packages []vpkg.Package) []string {
	candidates := make([]string, 0, len(packages))
	for _, pkg := range packages {
		candidates = append(candidates, pkg.Path)
	}
	sort.Strings(candidates)
	return candidates
}

func joinCandidates(candidates []string) string {
	return strings.Join(candidates, ", ")
}
