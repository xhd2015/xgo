//go:build go1.20
// +build go1.20

package syntax

import "cmd/compile/internal/syntax"

func (ctx *BlockContext) traverseCallStmtCallExpr(node syntax.Expr, globaleNames map[string]*DeclInfo, imports map[string]string) syntax.Expr {
	return ctx.traverseExpr(node, globaleNames, imports)
}
