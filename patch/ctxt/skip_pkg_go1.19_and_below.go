//go:build go1.17 && !go1.20
// +build go1.17,!go1.20

package ctxt

import (
	"strings"
)

func isSkippableSpecialPkg(pkgPath string) bool {
	if strings.HasPrefix(pkgPath, "golang.org/x/") {
		return true
	}
	return false
}
