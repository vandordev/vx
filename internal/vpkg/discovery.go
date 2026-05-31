package vpkg

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type Package struct {
	Path         string
	Root         string
	ManifestPath string
	Manifest     Manifest
}

func Discover(root string) ([]Package, error) {
	vpkgRoot := filepath.Join(root, "vpkg")

	info, err := os.Stat(vpkgRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	var packages []Package
	err = filepath.WalkDir(vpkgRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "vpkg.yaml" {
			return nil
		}

		manifest, err := LoadManifest(path)
		if err != nil {
			return err
		}

		packageRoot := filepath.Dir(path)
		relPath, err := filepath.Rel(vpkgRoot, packageRoot)
		if err != nil {
			return err
		}

		packages = append(packages, Package{
			Path:         filepath.ToSlash(relPath),
			Root:         packageRoot,
			ManifestPath: path,
			Manifest:     manifest,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Path < packages[j].Path
	})

	return packages, nil
}
