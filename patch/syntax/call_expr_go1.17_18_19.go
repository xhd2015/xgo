//go:build !go1.20
// +build !go1.20

package syntax

import "cmd/compile/internal/syntax"

func (ctx *BlockContext) traverseCallExpr(node *syntax.CallExpr, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.CallExpr {
	if node == nil {
		return nil
	}
	for i, arg := range node.ArgList {
		node.ArgList[i] = ctx.traverseExpr(arg, globaleNames, imports)
	}
	return node
}
