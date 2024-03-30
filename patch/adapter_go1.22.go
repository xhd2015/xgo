//go:build go1.22 && !go1.23
// +build go1.22,!go1.23

package patch

import (
	"go/constant"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

const goMajor = 1
const goMinor = 22

const genericTrapNeedsWorkaround = true
const closureMayBeEleminatedDueToIfConst = false

func forEachFunc(callback func(fn *ir.Func) bool) {
	for _, fn := range typecheck.Target.Funcs {
		if !callback(fn) {
			return
		}
	}
}

// go 1.20 does not require type
func NewNilExpr(pos src.XPos, t *types.Type) *ir.NilExpr {
	return ir.NewNilExpr(pos, t)
}

func AddFuncs(fn *ir.Func) {
	typecheck.Target.Funcs = append(typecheck.Target.Funcs, fn)
}

func NewFunc(fpos, npos src.XPos, sym *types.Sym, typ *types.Type) *ir.Func {
	return ir.NewFunc(fpos, npos, sym, typ)
}

func NewSignature(pkg *types.Pkg, recv *types.Field, tparams, params, results []*types.Field) *types.Type {
	return types.NewSignature(recv, params, results)
}

func NewBasicLit(pos src.XPos, t *types.Type, val constant.Value) ir.Node {
	return ir.NewBasicLit(pos, t, val)
}

// NOTE: []*types.Field instead of *types.Type(go1.21)
func takeAddrs(fn *ir.Func, t []*types.Field, nameOnly bool) ir.Expr {
	if len(t) == 0 {
		return NewNilExpr(fn.Pos(), intfSlice)
	}
	paramList := make([]ir.Node, len(t))
	i := 0
	ForEachField(t, func(field *types.Field) bool {
		paramList[i] = takeAddr(fn, field, nameOnly)
		i++
		return true
	})
	return ir.NewCompLitExpr(fn.Pos(), ir.OCOMPLIT, intfSlice, paramList)
}

// NOTE: []*types.Field instead of *types.Type(go1.21)
func getFieldNames(fn *ir.Func, t []*types.Field) ir.Expr {
	if len(t) == 0 {
		return NewNilExpr(fn.Pos(), strSlice)
	}
	paramList := make([]ir.Node, len(t))
	i := 0
	ForEachField(t, func(field *types.Field) bool {
		fieldName := getFieldName(fn, field)
		paramList[i] = NewStringLit(fn.Pos(), fieldName)
		i++
		return true
	})
	return ir.NewCompLitExpr(fn.Pos(), ir.OCOMPLIT, strSlice, paramList)
}

func getTypeNames(params []*types.Field) []ir.Node {
	paramNames := make([]ir.Node, 0, len(params))
	for _, p := range params {
		paramNames = append(paramNames, p.Nname.(*ir.Name))
	}
	return paramNames
}

func ForEachField(params []*types.Field, callback func(field *types.Field) bool) {
	n := len(params)
	for i := 0; i < n; i++ {
		if !callback(params[i]) {
			return
		}
	}
}

func GetFieldIndex(fields []*types.Field, i int) *types.Field {
	return fields[i]
}

func getCallee(callNode *ir.CallExpr) ir.Node {
	return callNode.Fun
}

// TODO: maybe go1.22 does not need this
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

func NewNameAt(pos src.XPos, sym *types.Sym, typ *types.Type) *ir.Name {
	return ir.NewNameAt(pos, sym, typ)
}

func isClosureWrapperForGeneric(fn *ir.Func) bool {
	return false
}
