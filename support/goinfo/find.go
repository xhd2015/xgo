package goinfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
)

func FindGoModDir(dir string) (string, error) {
	_, dir, err := FindGoModDirSubPath(dir)
	return dir, err
}

// FindGoModDirSubPath find go.mod by traversing dir bottom up,
// upon finding a go.mod file, it returns the corresponding
// sub path and root dir
func FindGoModDirSubPath(dir string) ([]string, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}

	// invariant: len(subPath) == level
	var subPath []string
	for level := 0; ; level++ {
		if level > 0 {
			baseName := filepath.Base(absDir)
			parent := filepath.Dir(absDir)
			if parent == absDir || parent == "" {
				return nil, "", fmt.Errorf("%s outside go module", dir)
			}
			absDir = parent
			subPath = append(subPath, baseName)
		}
		stat, err := os.Stat(filepath.Join(absDir, "go.mod"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, "", err
		}
		if stat.IsDir() {
			continue
		}
		return subPath, absDir, nil
	}
}

func GetModPath(dir string) (string, error) {
	output, err := cmd.Dir(dir).Output("go", "mod", "edit", "-json")
	if err != nil {
		return "", err
	}
	var goMod struct {
		Module struct {
			Path string
		}
	}
	err = json.Unmarshal([]byte(output), &goMod)
	if err != nil {
		return "", err
	}
	return goMod.Module.Path, nil
}
