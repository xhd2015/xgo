//go:build go1.18
// +build go1.18

package ast

import "go/ast"

func extractIndexListExpr(typeExpr ast.Expr) (ast.Expr, bool) {
	indexListExpr, ok := typeExpr.(*ast.IndexListExpr)
	if !ok {
		return typeExpr, false
	}
	return indexListExpr.X, true
}
