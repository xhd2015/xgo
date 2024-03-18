//go:build go1.17 && !go1.20
// +build go1.17,!go1.20

package syntax

import (
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	"strings"
)

func isSkippableSpecialPkg() bool {
	curPkgPath := xgo_ctxt.GetPkgPath()
	if strings.HasPrefix(curPkgPath, "golang.org/x/") {
		return true
	}
	return false
}
