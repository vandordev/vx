package vpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Manifest struct {
	APIVersion string            `yaml:"apiVersion"`
	Name       string            `yaml:"name"`
	Version    string            `yaml:"version"`
	Kind       string            `yaml:"kind"`
	Exports    map[string]Export `yaml:"exports"`
}

type Export struct {
	Kind        string           `yaml:"kind"`
	Templates   []TemplateRef    `yaml:"templates"`
	Files       []PathRef        `yaml:"files"`
	Directories []PathRef        `yaml:"directories"`
	Steps       []map[string]any `yaml:"steps"`
}

type TemplateRef struct {
	Path string `yaml:"path"`
}

type PathRef struct {
	Path string `yaml:"path"`
}

var (
	validPackageKinds = map[string]struct{}{
		"template-pack": {},
		"preset":        {},
		"tooling":       {},
	}
	validExportKinds = map[string]struct{}{
		"template": {},
		"preset":   {},
		"task":     {},
	}
)

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}

	packageRoot := filepath.Dir(path)
	if err := validateManifest(manifest, packageRoot); err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func validateManifest(manifest Manifest, packageRoot string) error {
	if strings.TrimSpace(manifest.APIVersion) == "" {
		return fmt.Errorf("manifest apiVersion is required")
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("manifest name is required")
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("manifest version is required")
	}
	if strings.TrimSpace(manifest.Kind) == "" {
		return fmt.Errorf("manifest kind is required")
	}
	if len(manifest.Exports) == 0 {
		return fmt.Errorf("manifest exports are required")
	}
	if _, ok := validPackageKinds[manifest.Kind]; !ok {
		return fmt.Errorf("manifest kind %q is invalid", manifest.Kind)
	}

	expectedName, err := expectedManifestName(packageRoot)
	if err != nil {
		return err
	}
	if manifest.Name != expectedName {
		return fmt.Errorf("manifest name %q does not match package path %q", manifest.Name, expectedName)
	}

	for exportName, export := range manifest.Exports {
		if strings.TrimSpace(export.Kind) == "" {
			return fmt.Errorf("export %q kind is required", exportName)
		}
		if _, ok := validExportKinds[export.Kind]; !ok {
			return fmt.Errorf("export %q kind %q is invalid", exportName, export.Kind)
		}
		if export.Kind != "template" {
			continue
		}
		if len(export.Templates) == 0 {
			return fmt.Errorf("export %q templates are required", exportName)
		}
		for _, template := range export.Templates {
			if err := validateTemplatePath(packageRoot, template.Path); err != nil {
				return fmt.Errorf("export %q template path %q: %w", exportName, template.Path, err)
			}
		}
	}

	return nil
}

func expectedManifestName(packageRoot string) (string, error) {
	segments := splitPath(packageRoot)
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "vpkg" {
			continue
		}
		if len(segments)-i < 3 {
			break
		}
		return filepath.ToSlash(filepath.Join(segments[i+1], segments[i+2])), nil
	}
	return "", fmt.Errorf("manifest path %q is not under vpkg/<namespace>/<package>", packageRoot)
}

func validateTemplatePath(packageRoot, rawPath string) error {
	if strings.TrimSpace(rawPath) == "" {
		return fmt.Errorf("must not be empty")
	}
	if filepath.IsAbs(rawPath) {
		return fmt.Errorf("must be relative to package root")
	}

	cleanPath := filepath.Clean(rawPath)
	if cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("must stay within package root")
	}
	if filepath.Ext(cleanPath) != ".vxt" {
		return fmt.Errorf("must point to a .vxt file")
	}

	resolvedPath := filepath.Join(packageRoot, cleanPath)
	rel, err := filepath.Rel(packageRoot, resolvedPath)
	if err != nil {
		return fmt.Errorf("resolve package root: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("must stay within package root")
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return fmt.Errorf("referenced file does not exist")
	}
	if info.IsDir() {
		return fmt.Errorf("must point to a file")
	}

	return nil
}

func splitPath(path string) []string {
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	clean = strings.TrimPrefix(clean, volume)
	clean = strings.TrimPrefix(clean, string(filepath.Separator))
	if clean == "" {
		return nil
	}
	return strings.Split(clean, string(filepath.Separator))
}
