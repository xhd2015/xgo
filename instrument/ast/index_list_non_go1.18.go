//go:build !go1.18
// +build !go1.18

package ast

import "go/ast"

func extractIndexListExpr(typeExpr ast.Expr) (ast.Expr, bool) {
	return typeExpr, false
}
