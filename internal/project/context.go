package project

import (
	"fmt"
	"path/filepath"
)

type Context struct {
	Root     string
	Language string
	Go       *GoContext
}

type GoContext struct {
	Module     string
	ModuleRoot string
}

func DetectContext(projectRoot string, cwd string) (Context, error) {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return Context{}, fmt.Errorf("resolve project root: %w", err)
	}
	return Context{Root: absRoot}, nil
}
