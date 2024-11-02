package git

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

func MustCatFile(dir string, ref string, file string) (content string, err error) {
	var ok bool
	ok, content, err = CatFile(dir, ref, file)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("%s does not exist in %s", file, ref)
		return
	}
	return
}
func CatFile(dir string, ref string, file string) (ok bool, content string, err error) {
	if ref == "" {
		err = fmt.Errorf("requires ref")
		return
	}
	if file == "" {
		err = fmt.Errorf("requires file")
		return
	}
	if ref == COMMIT_WORKING {
		contentBytes, fileErr := os.ReadFile(path.Join(dir, file))
		if fileErr != nil {
			if os.IsNotExist(fileErr) {
				return
			}
			err = fileErr
			return
		}
		ok = true
		content = string(contentBytes)
		return
	}
	var stderrBuf strings.Builder
	content, err = cmd.Dir(dir).Stderr(&stderrBuf).Output("git", "cat-file", "-p", ref+":"+file)
	stderr := stderrBuf.String()
	// example
	//    fatal: path 'go.mod2' does not exist in 'master'
	// example2:
	//    exists on disk, but not in
	ok = true
	if strings.Contains(stderr, "does not exist in") || !hasFile(dir, ref, file) {
		ok = false
		err = nil
		return
	}

	return
}

// see: https://stackoverflow.com/questions/18461761/git-check-whether-file-exists-in-some-version
func hasFile(dir string, ref string, file string) bool {
	err := cmd.Dir(dir).Run("git", "cat-file", "-e", ref+":"+file)
	return err == nil
}
