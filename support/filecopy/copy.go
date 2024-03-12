package filecopy

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CopyReplaceDir(srcDir string, targetDir string, useLink bool) error {
	if srcDir == "" {
		return fmt.Errorf("requires srcDir")
	}
	targetAbsDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	if targetAbsDir == "/" {
		return fmt.Errorf("cannot replace /")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(targetAbsDir, homeDir+"/.xgo") && !strings.HasPrefix(targetAbsDir, "/tmp/") {
		return fmt.Errorf("replace not permitted:%s", targetDir)
	}
	err = os.RemoveAll(targetAbsDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(targetAbsDir), 0755)
	if err != nil {
		return err
	}
	if useLink {
		return LinkFiles(srcDir, targetAbsDir)
	}
	// TODO: use native go
	return exec.Command("cp", "-R", srcDir, targetAbsDir).Run()
}

// NOTE: sym link must use abs path to ensure the file work correctly
func LinkFiles(srcDir string, targetDir string) error {
	absDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}
	return filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, absDir) {
			return fmt.Errorf("invalid path: %s", path)
		}
		relPath := path[len(absDir):]
		if strings.HasPrefix(relPath, string(os.PathSeparator)) {
			relPath = relPath[1:]
		}
		// if relPath is "", it is root dir
		targetPath := filepath.Join(targetDir, relPath)
		if d.IsDir() {
			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				return err
			}
			return nil
		}
		// link file
		return os.Symlink(path, targetPath)
	})
}
func LinkFile(srcFile string, dstFile string) error {
	absFile, err := filepath.Abs(srcFile)
	if err != nil {
		return err
	}
	return os.Symlink(absFile, dstFile)
}

func LinkDirToTmp(srcDir string, tmpDir string) (dstTmpDir string, err error) {
	subTmp, err := os.MkdirTemp(tmpDir, filepath.Base(srcDir))
	if err != nil {
		return "", err
	}
	err = LinkFiles(srcDir, subTmp)
	if err != nil {
		return "", err
	}
	return subTmp, nil
}
