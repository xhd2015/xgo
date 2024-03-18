package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/fileutil"
)

// fixup src dir to prepare for release build
func fixupSrcDir(targetDir string, rev string) error {
	err := switchRelease(targetDir)
	if err != nil {
		return err
	}
	err = embedPatchCompiler(targetDir)
	if err != nil {
		return err
	}
	err = updateRevisions(targetDir, false, rev)
	if err != nil {
		return err
	}
	return nil
}

func switchRelease(targetDir string) error {
	devFile := filepath.Join(targetDir, "cmd", "xgo", "env_dev.go")
	err := fileutil.Patch(devFile, func(data []byte) ([]byte, error) {
		content := string(data)
		content = "//go:build ignore\n" + content
		return []byte(content), nil
	})
	if err != nil {
		return err
	}
	releaseFile := filepath.Join(targetDir, "cmd", "xgo", "env_release.go")
	err = fileutil.Patch(releaseFile, func(data []byte) ([]byte, error) {
		buildIgnore := "//go:build ignore"
		content := string(data)
		idx := strings.Index(content, buildIgnore)
		if idx < 0 {
			fmt.Printf("content: %s\n", content)
			return nil, fmt.Errorf("missing %s in %s", buildIgnore, releaseFile)
		}
		content = strings.ReplaceAll(content, buildIgnore, "")
		return []byte(content), nil
	})
	if err != nil {
		return err
	}
	return nil
}

func embedPatchCompiler(targetDir string) error {
	patchCompiler := filepath.Join(targetDir, "cmd", "xgo", "patch_compiler")
	err := os.Rename(filepath.Join(targetDir, "patch"), patchCompiler)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(patchCompiler, "go.mod"), filepath.Join(patchCompiler, "go.mod.txt"))
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(targetDir, "runtime", "trap_runtime"), filepath.Join(patchCompiler, "trap_runtime"))
	if err != nil {
		return err
	}
	return nil
}
func updateRevisions(targetDir string, unlink bool, rev string) error {
	// unlink files because all files are symlink
	files := revision.GetVersionFiles(targetDir)
	if unlink {
		for _, file := range files {
			err := unlinkFile(file)
			if err != nil {
				return err
			}
		}
	}

	for _, file := range files {
		err := revision.PatchVersionFile(file, rev)
		if err != nil {
			return err
		}
	}
	return nil
}
