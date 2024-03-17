//go:build go1.18 && !go1.19
// +build go1.18,!go1.19

package patch

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
)

const genericTrapNeedsWorkaround = true

func NewSignature(pkg *types.Pkg, recv *types.Field, tparams, params, results []*types.Field) *types.Type {
	return types.NewSignature(pkg, recv, tparams, params, results)
}

func SetConvTypeWordPtr(conv *ir.ConvExpr, t *types.Type) {

}

func getFuncResultsType(funcType *types.Type) *types.Type {
	return funcType.FuncType().Results
}

func typeNode(t *types.Type) ir.Ntype {
	return ir.TypeNode(t)
}

// go1.18 workaround, slice needs to have type
func wrapListType(expr *ir.CompLitExpr) *ir.CompLitExpr {
	expr.SetType(intfSlice)
	return expr
}
