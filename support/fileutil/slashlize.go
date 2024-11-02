package fileutil

import (
	"path/filepath"
	"strings"
)

func Slashlize(path string) string {
	if filepath.Separator == '/' {
		return path
	}
	return strings.ReplaceAll(path, string(filepath.Separator), "/")
}
