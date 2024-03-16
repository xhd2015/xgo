package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
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

func setupDevDir(goroot string, printDir bool) error {
	cmdCompile := filepath.Join(goroot, "src", "cmd", "compile")
	vscodeDir := filepath.Join(cmdCompile, ".vscode")

	err := assertDir(cmdCompile)
	if err != nil {
		return fmt.Errorf("setup $GOROOT/src/cmd/compile: %w", err)
	}
	err = os.MkdirAll(vscodeDir, 0755)
	if err != nil {
		return err
	}
	if printDir {
		fmt.Printf("%s\n", cmdCompile)
	}
	// {
	// 	"go.goroot": "$GOROOT",
	// 	"gopls": {
	// 		"build.env": {
	// 			"GOROOT": "$GOROOT",
	// 			"PATH": "$GOROOT/bin:${env:PATH}"
	// 		}
	// 	}
	// }
	settingsFile := filepath.Join(vscodeDir, "settings.json")
	err = fileutil.PatchJSONPretty(settingsFile, func(settings *map[string]interface{}) error {
		if *settings == nil {
			*settings = make(map[string]interface{})
		}
		// NOTE: must use absolute path, not rel path
		// gorootRel := "../.."
		gorootRel := goroot
		(*settings)["go.goroot"] = gorootRel
		gopls, _ := (*settings)["gopls"].(map[string]interface{})
		if gopls == nil {
			gopls = make(map[string]interface{})
			(*settings)["gopls"] = gopls
		}
		buildEnv, _ := gopls["build.env"].(map[string]interface{})
		if buildEnv == nil {
			buildEnv = make(map[string]interface{})
			gopls["build.env"] = buildEnv
		}
		buildEnv["GOROOT"] = gorootRel
		buildEnv["PATH"] = fmt.Sprintf("%s/bin:${env:PATH}", gorootRel)
		return nil
	})

	if err != nil {
		return fmt.Errorf("updating %s: %w", settingsFile, err)
	}
	return nil
}
