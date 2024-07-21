//go:build !go1.18
// +build !go1.18

package astdiff

import "go/ast"

func funcTypeSame(a *ast.FuncType, b *ast.FuncType) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if !fieldListSame(a.Params, b.Params) {
		return false
	}
	if !fieldListSame(a.Results, b.Results) {
		return false
	}
	return true
}
