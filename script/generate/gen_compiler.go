package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/filecopy"
)

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
