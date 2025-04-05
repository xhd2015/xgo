//go:build !go1.18
// +build !go1.18

package ast

import (
	"go/ast"
)

func IsGenericFunc(funcDecl *ast.FuncDecl) bool {
	return false
}
