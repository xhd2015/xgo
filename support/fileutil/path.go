package fileutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

func CleanSpecial(path string) string {
	chars := []rune(path)
	n := len(chars)
	j := 0
	for i := 0; i < n; i++ {
		ch := chars[i]
		if ch == ' ' {
			chars[j] = '_'
			j++
			continue
		}
		if ch < 128 && !(ch == '.' || ch == '_' || ch == '-' || (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			continue
		}
		chars[j] = chars[i]
		j++
	}
	return string(chars[:j])
}

// when doing filepath.Join(a,b),
// on windows, if b has :, everything fails
// TODO: test on windows
func RebasePath(root string, path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return RebaseAbsPath(root, absPath)
}

func RebaseAbsPath(root string, absPath string) string {
	if runtime.GOOS != "windows" {
		return filepath.Join(root, absPath)
	}
	idx := strings.Index(absPath, ":")
	if idx < 0 {
		return filepath.Join(root, absPath)
	}
	// join two parts before and after :
	return filepath.Join(root, absPath[:idx], absPath[idx+1:])
}
