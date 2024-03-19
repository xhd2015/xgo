//go:build go1.19 && !go1.20
// +build go1.19,!go1.20

package patch

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
)

const goMajor = 1
const goMinor = 19

const genericTrapNeedsWorkaround = true

func NewSignature(pkg *types.Pkg, recv *types.Field, tparams, params, results []*types.Field) *types.Type {
	return types.NewSignature(pkg, recv, tparams, params, results)
}

func SetConvTypeWordPtr(conv *ir.ConvExpr, t *types.Type) {

}

func getFuncResultsType(funcType *types.Type) *types.Type {
	return funcType.FuncType().Results
}

func canInsertTrap(fn *ir.Func) bool {
	curPkgPath := xgo_ctxt.GetPkgPath()
	fnPkgPath := fn.Sym().Pkg.Path
	if curPkgPath != fnPkgPath {
		return false
	}
	if isSkippableSpecialPkg() {
		return false
	}
	// fnName := fn.Sym().Name
	// if strings.Contains(fnName, "[") && strings.Contains(fnName, "]") {
	// 	return false
	// }
	return true
}
