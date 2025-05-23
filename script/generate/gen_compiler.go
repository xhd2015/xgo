package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
)

func copyCompilerInstrument(rootDir string, genName string) error {
	err := copyCompilerExtra(rootDir, genName)
	if err != nil {
		return err
	}

	err = copyCompilerConstants(rootDir, genName)
	if err != nil {
		return err
	}

	return nil
}

func copyCompilerExtra(rootDir string, genName string) error {
	srcDir := filepath.Join(rootDir, "instrument", "compiler_extra")
	genDir := filepath.Join(rootDir, "patch", "instrument", "compiler_extra")
	err := filecopy.CopyReplaceDir(srcDir, genDir, false)
	if err != nil {
		return err
	}

	return prependGeneratePrelude(genDir, genName, func(path, relPath string, content []byte) []byte {
		if relPath == "register.go" {
			// replace
			// "github.com/xhd2015/xgo/instrument/constants"
			// "cmd/compile/internal/xgo_rewrite_internal/patch/instrument/constants"
			content = bytes.ReplaceAll(content, []byte(`"github.com/xhd2015/xgo/instrument/constants"`), []byte(`"cmd/compile/internal/xgo_rewrite_internal/patch/instrument/constants"`))
		}
		return content
	})
}

func prependGeneratePrelude(dir string, genName string, process func(path string, relPath string, content []byte) []byte) error {
	prelude := getPrelude(genName)
	return fileutil.WalkRelative(dir, func(path string, relPath string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if process != nil {
			content = process(path, relPath, content)
		}
		content = []byte(prelude + string(content))
		err = os.WriteFile(path, content, 0755)
		if err != nil {
			return err
		}
		return nil
	})
}

func prependGeneratePreludeFile(file string, genName string) error {
	prelude := getPrelude(genName)
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	content = []byte(prelude + string(content))
	return os.WriteFile(file, content, 0755)
}

func getPrelude(genName string) string {
	return fmt.Sprintf("// Code generated by `go run ./script/generate %s`. DO NOT EDIT.\n\n", genName)
}

func copyCompilerConstants(rootDir string, genName string) error {
	srcDir := filepath.Join(rootDir, "instrument", "constants")
	genDir := filepath.Join(rootDir, "patch", "instrument", "constants")
	err := filecopy.NewOptions().Ignore("constants.go").CopyReplaceDir(srcDir, genDir)
	if err != nil {
		return err
	}

	return prependGeneratePrelude(genDir, genName, nil)
}

func genXgoCompilerPatch(rootDir string) error {
	patchDir := filepath.Join(rootDir, "patch")
	genPatchDir := filepath.Join(rootDir, "cmd", "xgo", "asset", "compiler_patch_gen")

	return copyGoModule(patchDir, genPatchDir, []string{".xgo", "test", "legacy"})
}

func generateCompilerPatch(rootDir string) error {
	if true {
		return nil
	}
	srcDir := filepath.Join(rootDir, "patch")
	genDir := filepath.Join(rootDir, "cmd", "xgo", "patch_compiler_gen")
	genFile := genDir + ".go"

	err := filecopy.CopyReplaceDir(srcDir, genDir, false)
	if err != nil {
		return err
	}

	err = os.Remove(filepath.Join(genDir, "go.mod"))
	if err != nil {
		return err
	}

	// rename all .go to .go.txt
	err = filepath.WalkDir(genDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		return os.Rename(path, path+".txt")
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(genFile, []byte(prelude+`package main\n\n`+"//go:generate ../../scrip/generate "+string(GenernateType_CompilerPatch)), 0755)
	if err != nil {
		return err
	}

	return nil
}
