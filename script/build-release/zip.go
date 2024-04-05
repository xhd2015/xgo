package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func createInstallSRC(zipFile string, projectRoot string) error {
	upgradeDir := filepath.Join(projectRoot, "script", "install")
	return createZip(zipFile, upgradeDir)
}

func createZip(zipFile string, dir string) error {
	archive, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()
	if dir == "" {
		return fmt.Errorf("requires dir")
	}
	// trim last
	if dir[len(dir)-1] == filepath.Separator {
		dir = dir[:len(dir)-1]
	}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath := strings.TrimPrefix(path, dir)
		relPath = relPath[1:] // remove head separator

		if filepath.Separator != '/' {
			relPath = strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		}
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		w, err := zipWriter.Create(relPath)
		if err != nil {
			return fmt.Errorf("adding %s: %w", relPath, err)
		}
		_, err = io.Copy(w, r)
		return err
	})
}
