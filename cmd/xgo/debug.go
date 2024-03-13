package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getVscodeDebugFile(tmpDir string, vscode string) (vscodeDebugFile string, suffix string, err error) {
	const stdoutParamPrefix = "stdout?"
	if vscode == "" {
		f, err := os.Stat(".vscode")
		if err == nil && f.IsDir() {
			vscodeDebugFile = filepath.Join(".vscode", "launch.json")
		}
	} else if vscode == "stdout" || strings.HasPrefix(vscode, stdoutParamPrefix) {
		vscodeDebugFile = filepath.Join(tmpDir, "vscode_launch.json")
	} else {
		f, err := os.Stat(vscode)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return "", "", fmt.Errorf("check vscode debug flag: %w", err)

			}
			err = nil
			// treat as file
		}
		if f != nil && f.IsDir() {
			vscodeDebugFile = filepath.Join(vscode, "launch.json")
		} else {
			vscodeDebugFile = vscode
		}
	}
	// make abs
	if vscodeDebugFile != "" {
		var err error
		vscodeDebugFile, err = filepath.Abs(vscodeDebugFile)
		if err != nil {
			return "", "", err
		}
	}
	if strings.HasPrefix(vscode, stdoutParamPrefix) {
		suffix = "?" + strings.TrimPrefix(vscode, stdoutParamPrefix)
	}
	return vscodeDebugFile, suffix, nil
}
