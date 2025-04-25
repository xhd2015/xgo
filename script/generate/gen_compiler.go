package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/filecopy"
)

func copyCompilerInstrument(rootDir string) error {
	err := copyCompilerExtra(rootDir)
	if err != nil {
		return err
	}

	err = copyCompilerConstants(rootDir)
	if err != nil {
		return err
	}

	return nil
}

func copyCompilerExtra(rootDir string) error {
	srcDir := filepath.Join(rootDir, "instrument", "compiler_extra")
	genDir := filepath.Join(rootDir, "patch", "instrument", "compiler_extra")
	err := filecopy.CopyReplaceDir(srcDir, genDir, false)
	if err != nil {
		return err
	}
	// replace
	// "github.com/xhd2015/xgo/instrument/constants"
	// "cmd/compile/internal/xgo_rewrite_internal/patch/instrument/constants"
	registerFile := filepath.Join(genDir, "register.go")
	content, err := os.ReadFile(registerFile)
	if err != nil {
		return err
	}
	content = bytes.ReplaceAll(content, []byte(`"github.com/xhd2015/xgo/instrument/constants"`), []byte(`"cmd/compile/internal/xgo_rewrite_internal/patch/instrument/constants"`))

	err = os.WriteFile(registerFile, content, 0755)
	if err != nil {
		return err
	}
	return nil
}

func copyCompilerConstants(rootDir string) error {
	srcDir := filepath.Join(rootDir, "instrument", "constants")
	genDir := filepath.Join(rootDir, "patch", "instrument", "constants")
	return filecopy.NewOptions().Ignore("constants.go").CopyReplaceDir(srcDir, genDir)
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
