package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrRootNotFound = errors.New("vx project root not found")

func FindRoot(start string) (string, error) {
	if start == "" {
		start = "."
	}

	current, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve project start path: %w", err)
	}

	info, err := os.Stat(current)
	if err != nil {
		return "", fmt.Errorf("stat project start path: %w", err)
	}
	if !info.IsDir() {
		current = filepath.Dir(current)
	}

	for {
		vpkgPath := filepath.Join(current, "vpkg")
		if info, err := os.Stat(vpkgPath); err == nil && info.IsDir() {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("%w: %s", ErrRootNotFound, start)
		}
		current = parent
	}
}
