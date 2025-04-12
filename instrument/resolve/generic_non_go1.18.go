//go:build !go1.18
// +build !go1.18

package resolve

import (
	"go/ast"

	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// something[A,B]
func (c *Scope) traverseIndexListExpr(expr ast.Expr) bool {
	return false
}

func (c *Scope) resolveIndexListExpr(expr ast.Expr) (types.Info, bool) {
	return nil, false
}
