package filecopy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
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
	// use native go
	return copyDir(srcDir, targetAbsDir)
	// return exec.Command("cp", "-R", srcDir, targetAbsDir).Run()
}

func copyDir(srcDir string, targetAbsDir string) error {
	n := len(srcDir)
	prefixLen := n + len(string(filepath.Separator))
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// root
		if path == srcDir {
			return os.MkdirAll(targetAbsDir, 0755)
		}
		subPath := path[prefixLen:]
		dstPath := filepath.Join(targetAbsDir, subPath)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		reader, err := os.Open(path)
		if err != nil {
			return err
		}
		defer reader.Close()

		writer, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		return err
	})
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
