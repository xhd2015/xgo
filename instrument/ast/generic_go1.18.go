//go:build go1.18
// +build go1.18

package ast

import (
	"go/ast"
)

func IsGenericFunc(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.TypeParams == nil || len(funcDecl.Type.TypeParams.List) == 0 {
		return false
	}
	return true
}
