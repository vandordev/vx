package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return Context{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	if info, err := os.Stat(absCWD); err == nil && !info.IsDir() {
		absCWD = filepath.Dir(absCWD)
	}

	ctx := Context{Root: absRoot}
	goCtx, err := detectGoContext(absRoot, absCWD)
	if err != nil {
		return Context{}, err
	}
	if goCtx != nil {
		ctx.Language = "go"
		ctx.Go = goCtx
	}
	return ctx, nil
}

func detectGoContext(projectRoot string, cwd string) (*GoContext, error) {
	current := cwd
	for {
		if !isWithinRoot(projectRoot, current) {
			return nil, nil
		}

		goModPath := filepath.Join(current, "go.mod")
		if info, err := os.Stat(goModPath); err == nil && !info.IsDir() {
			module, err := parseGoModule(goModPath)
			if err != nil {
				return nil, err
			}
			if module == "" {
				return nil, nil
			}

			moduleRoot, err := filepath.Rel(projectRoot, current)
			if err != nil {
				return nil, fmt.Errorf("resolve go module root: %w", err)
			}
			return &GoContext{
				Module:     module,
				ModuleRoot: moduleRoot,
			}, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return nil, nil
		}
		current = parent
	}
}

func parseGoModule(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read go module file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "module" {
			continue
		}

		module := fields[1]
		if unquoted, err := strconv.Unquote(module); err == nil {
			module = unquoted
		}
		return module, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan go module file %q: %w", path, err)
	}

	return "", nil
}

func isWithinRoot(root string, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
