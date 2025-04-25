//go:build !go1.19
// +build !go1.19

package ctxt

import (
	"cmd/compile/internal/base"
)

// with go1.18, types.LocalPkg.Path is empty
func GetPkgPath() string {
	return base.Ctxt.Pkgpath
}
