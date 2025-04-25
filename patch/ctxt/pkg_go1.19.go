//go:build go1.19
// +build go1.19

package ctxt

import "cmd/compile/internal/types"

func GetPkgPath() string {
	return types.LocalPkg.Path
}
