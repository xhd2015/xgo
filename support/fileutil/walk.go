package fileutil

import (
	"io/fs"
	"path/filepath"
)

// WalkRelative: calculate relative path when walking
func WalkRelative(root string, h func(path string, relPath string, d fs.DirEntry) error) error {
	cleanRoot := filepath.Clean(root)
	n := len(cleanRoot)
	prefixLen := n + len(string(filepath.Separator))
	return filepath.WalkDir(cleanRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// root
		if path == cleanRoot {
			return h(path, "", d)
		}
		subPath := path[prefixLen:]

		return h(path, subPath, d)
	})
}
