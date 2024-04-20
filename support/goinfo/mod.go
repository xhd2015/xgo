package goinfo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/osinfo"
)

var ErrGoModNotFound = errors.New("go.mod not found")
var ErrGoModDoesNotHaveModule = errors.New("go.mod does not have module")

func ResolveMainModule(dir string, args []string) (subPaths []string, mainModule string, err error) {
	goMod, subPaths, err := findGoMod(dir)
	if err != nil {
		return nil, "", err
	}

	goModContent, err := os.ReadFile(goMod)
	if err != nil {
		return nil, "", err
	}
	modPath := parseModPath(string(goModContent))
	if modPath == "" {
		return nil, "", ErrGoModDoesNotHaveModule
	}

	return subPaths, modPath, nil
}

func isRelative(arg string) bool {
	if arg == "" {
		// pwd
		return true
	}
	n := len(arg)
	if arg[0] != '.' {
		return false
	}
	if n == 1 || arg[1] == '/' || (osinfo.IS_WINDOWS && arg[1] == '\\') {
		// . ./ .\
		return true
	}
	if arg[1] != '.' {
		return false
	}
	return n == 2 || arg[2] == '/' || (osinfo.IS_WINDOWS && arg[2] == '\\')
}

func findGoMod(dir string) (file string, subPaths []string, err error) {
	var absDir string
	if dir == "" {
		absDir, err = os.Getwd()
	} else {
		absDir, err = filepath.Abs(dir)
	}
	if err != nil {
		return "", nil, err
	}
	iterDir := absDir
	init := true
	for {
		if init {
			init = false
		} else {
			subPaths = append(subPaths, filepath.Base(iterDir))
			nextIterDir := filepath.Dir(iterDir)
			if iterDir == string(filepath.Separator) || nextIterDir == iterDir {
				// until root
				// TODO: what about windows?
				return "", nil, ErrGoModNotFound
			}
			iterDir = nextIterDir
		}
		file := filepath.Join(iterDir, "go.mod")
		stat, err := os.Stat(file)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return "", nil, err
			}
			continue
		}
		if stat.IsDir() {
			continue
		}
		// a valid go.mod found
		return file, subPaths, nil
	}
}

func parseModPath(goModContent string) string {
	lines := strings.Split(string(goModContent), "\n")
	n := len(lines)
	for i := 0; i < n; i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimSpace(line[len("module "):])
			commentIdx := strings.Index(module, "//")
			if commentIdx >= 0 {
				module = strings.TrimSpace(module[:commentIdx])
			}
			return module
		}
	}
	return ""
}
