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
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	var init bool = true
	for {
		if !init {
			parent := filepath.Dir(absDir)
			if parent == absDir || parent == "" {
				return "", fmt.Errorf("%s outside go module", dir)
			}
			absDir = parent
		}
		stat, err := os.Stat(filepath.Join(absDir, "go.mod"))
		init = false
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", err
		}
		if stat.IsDir() {
			continue
		}
		return absDir, nil
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
