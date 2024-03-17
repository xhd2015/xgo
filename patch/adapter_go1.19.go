//go:build go1.19 && !go1.20
// +build go1.19,!go1.20

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
