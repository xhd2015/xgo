package embed

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CopyOptions struct {
	IgnorePaths []string
}

func CopyDir(srcFS embed.FS, subName string, dstDir string, opts CopyOptions) error {
	return fs.WalkDir(srcFS, subName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == subName {
			return os.MkdirAll(dstDir, 0755)
		}

		name := path[len(subName)+len("/"):]
		if len(opts.IgnorePaths) > 0 {
			for _, ignorePath := range opts.IgnorePaths {
				if name == ignorePath {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}
		}
		// join without prefix `subName`, also works on windows
		dstPath := filepath.Join(dstDir, embedPathToFsPath(name))
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		content, err := srcFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, content, 0755)
	})
}

func embedPathToFsPath(embedPath string) string {
	if filepath.Separator == '/' {
		return embedPath
	}
	return strings.ReplaceAll(embedPath, "/", string(filepath.Separator))
}
