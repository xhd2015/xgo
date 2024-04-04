//go:build go1.17 && !go1.22
// +build go1.17,!go1.22

package patch

import (
	"go/constant"

	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
)

// for go1.20, target does not have Target.Funcs, instead, use Target.Decls
func forEachFunc(callback func(fn *ir.Func) bool) {
	// for go1.21 and above, this can just be:
	//   for _, fn := range typecheck.Target.Funcs
	for _, decl := range typecheck.Target.Decls {
		fn, ok := decl.(*ir.Func)
		if !ok {
			continue
		}
		if !callback(fn) {
			return
		}
	}
}

// go 1.20 does not require type
func NewNilExpr(pos src.XPos, t *types.Type) *ir.NilExpr {
	nilExpr := ir.NewNilExpr(pos)
	nilExpr.SetType(t)
	return nilExpr
}

func AddFuncs(fn *ir.Func) {
	// for go 1.21, use typecheck.Target.Funcs
	typecheck.Target.Decls = append(typecheck.Target.Decls, fn)
}

func NewFunc(fpos, npos src.XPos, sym *types.Sym, typ *types.Type) *ir.Func {
	// name := ir.NewNameAt(npos, sym, typ) // go1.22
	name := ir.NewNameAt(npos, sym)
	name.SetType(typ)
	name.Class = ir.PFUNC
	sym.SetFunc(true)

	// fn := &ir.Func{Nname: name} // go1.22
	fn := ir.NewFunc(fpos)
	fn.Nname = name
	// Most functions are ABIInternal. The importer or symabis
	// pass may override this.
	fn.ABI = obj.ABIInternal
	fn.SetTypecheck(1)

	name.Func = fn
	name.Defn = fn

	return fn
}

func NewBasicLit(pos src.XPos, t *types.Type, val constant.Value) ir.Node {
	lit := ir.NewBasicLit(pos, val)
	lit.SetType(t)
	return lit
}

func takeAddrs(fn *ir.Func, t *types.Type, nameOnly bool) ir.Expr {
	if t.NumFields() == 0 {
		return NewNilExpr(fn.Pos(), intfSlice)
	}
	paramList := make([]ir.Node, t.NumFields())
	i := 0
	ForEachField(t, func(field *types.Field) bool {
		paramList[i] = takeAddr(fn, field, nameOnly)
		i++
		return true
	})
	return wrapListType(ir.NewCompLitExpr(fn.Pos(), ir.OCOMPLIT, typeNode(intfSlice), paramList))
}

func getFieldNames(pos src.XPos, fn *ir.Func, t *types.Type) ir.Expr {
	if t.NumFields() == 0 {
		return NewNilExpr(pos, strSlice)
	}
	paramList := make([]ir.Node, t.NumFields())
	i := 0
	ForEachField(t, func(field *types.Field) bool {
		fieldName := getFieldName(fn, field)
		paramList[i] = NewStringLit(pos, fieldName)
		i++
		return true
	})
	return wrapListType(ir.NewCompLitExpr(pos, ir.OCOMPLIT, typeNode(strSlice), paramList))
}

func getTypeNames(params *types.Type) []ir.Node {
	n := params.NumFields()
	paramNames := make([]ir.Node, 0, n)
	for i := 0; i < n; i++ {
		p := params.Field(i)
		paramNames = append(paramNames, p.Nname.(*ir.Name))
	}
	return paramNames
}

func ForEachField(params *types.Type, callback func(field *types.Field) bool) {
	n := params.NumFields()
	for i := 0; i < n; i++ {
		if !callback(params.Field(i)) {
			return
		}
	}
}
func GetFieldIndex(fields *types.Type, i int) *types.Field {
	return fields.Field(i)
}

func getCallee(callNode *ir.CallExpr) ir.Node {
	return callNode.X
}

func NewNameAt(pos src.XPos, sym *types.Sym, typ *types.Type) *ir.Name {
	n := ir.NewNameAt(pos, sym)
	n.SetType(typ)
	return n
}

const closureMayBeEliminatedDueToIfConst = true
