//go:build go1.21 && !go1.22
// +build go1.21,!go1.22

package patch

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/types"
)

const goMajor = 1
const goMinor = 21

const genericTrapNeedsWorkaround = false

// different with go1.20
func NewSignature(pkg *types.Pkg, recv *types.Field, tparams, params, results []*types.Field) *types.Type {
	return types.NewSignature(recv, params, results)
}

func SetConvTypeWordPtr(conv *ir.ConvExpr, t *types.Type) {
	conv.TypeWord = reflectdata.TypePtrAt(base.Pos, types.NewPtr(t))
}

func SetConvTypeWord(conv *ir.ConvExpr, t *types.Type) {
	conv.TypeWord = reflectdata.TypePtrAt(base.Pos, t)
}

func getFuncResultsType(funcType *types.Type) *types.Type {
	panic("getFuncResultsType should not be called above go1.19")
}

func canInsertTrap(fn *ir.Func) bool {
	return true
}

func isClosureWrapperForGeneric(fn *ir.Func) bool {
	return false
}
