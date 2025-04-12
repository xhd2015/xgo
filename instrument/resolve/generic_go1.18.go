//go:build go1.18
// +build go1.18

package resolve

import (
	"go/ast"

	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// something[A,B]
func (c *Scope) traverseIndexListExpr(expr ast.Expr) bool {
	list, ok := expr.(*ast.IndexListExpr)
	if !ok {
		return false
	}
	c.traverseExpr(list.X)
	for _, index := range list.Indices {
		c.traverseExpr(index)
	}
	return true
}

func (c *Scope) resolveIndexListExpr(expr ast.Expr) (types.Info, bool) {
	list, ok := expr.(*ast.IndexListExpr)
	if !ok {
		return nil, false
	}
	info := c.resolveInfo(list.X)
	// if info is a Type, then this is a
	// generic type
	switch info := info.(type) {
	case types.Type:
		return info, true
	case types.PkgFunc:
		return info, true
	case types.PkgTypeMethod:
		return info, true
	}
	return types.Unknown{}, true
}
