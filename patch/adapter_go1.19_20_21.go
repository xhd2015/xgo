//go:build go1.19 && !go1.22
// +build go1.19,!go1.22

package patch

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
)

func typeNode(t *types.Type) *types.Type {
	return t
}

func wrapListType(expr *ir.CompLitExpr) *ir.CompLitExpr {
	return expr
}
