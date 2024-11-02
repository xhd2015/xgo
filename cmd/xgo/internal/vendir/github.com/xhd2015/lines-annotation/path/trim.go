package path

import (
	"fmt"
	"path/filepath"
	"strings"
)

// TrimPrefix (/a/b/c, /a/b) -> c
func TrimPrefix(s string, prefix string) (string, error) {
	return doTrimPathPrefix(s, prefix, false, string(filepath.Separator))
}

func TrimPrefixOrEmpty(s string, prefix string) string {
	res, _ := doTrimPathPrefix(s, prefix, true, string(filepath.Separator))
	return res
}

// empty indicates not matching prefix
func TrimModPrefixOrEmpty(pkgFile string, modPath string) string {
	res, _ := doTrimPathPrefix(pkgFile, modPath, true, "/")
	return res
}

func doTrimPathPrefix(s string, prefix string, allowUnprefixed bool, sep string) (string, error) {
	if !strings.HasPrefix(s, prefix) {
		if !allowUnprefixed {
			return "", fmt.Errorf("string %s not prefixed with %s", s, prefix)
		}
		return "", nil
	}
	s = s[len(prefix):]
	return strings.TrimPrefix(s, sep), nil
}
