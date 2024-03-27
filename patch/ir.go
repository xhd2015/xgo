package patch

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"go/constant"
	"runtime"
	"strings"
)

func ifConstant(pos src.XPos, b bool, body []ir.Node, els []ir.Node) *ir.IfStmt {
	return ir.NewIfStmt(pos,
		NewBoolLit(pos, true),
		body,
		els,
	)
}

func NewStringLit(pos src.XPos, s string) ir.Node {
	return NewBasicLit(pos, types.Types[types.TSTRING], constant.MakeString(s))
}

func NewInt64Lit(pos src.XPos, i int64) ir.Node {
	return NewBasicLit(pos, types.Types[types.TINT64], constant.MakeInt64(i))
}

func NewIntLit(pos src.XPos, i int) ir.Node {
	return NewBasicLit(pos, types.Types[types.TINT], constant.MakeInt64(int64(i)))
}

func NewBoolLit(pos src.XPos, b bool) ir.Node {
	return NewBasicLit(pos, types.Types[types.TBOOL], constant.MakeBool(b))
}

// how to delcare a new function?
// init names are usually init.0, init.1, ...
//
// NOTE: when there is already an init function, declare new init function
// will give an error: main..inittask: relocation target main.init.1 not defined

// with go1.22, modifying existing init's body may cause the error:
//
//	__xgo_autogen.go:2:6: internal compiler error: unexpected initialization statement: (.)
const needToCreateInit = goMajor > 1 || (goMajor == 1 && goMinor >= 22)

// Deprecated: use __xgo_link_generate_init_regs_body instead
// principal: Don't rely too much on IR
// don't add function on the fly, instead, modify existing
// functions body
func prependInit(pos src.XPos, target *ir.Package, body []ir.Node) {
	if !needToCreateInit && len(target.Inits) > 0 {
		target.Inits[0].Body.Prepend(body...)
		return
	}

	sym := types.LocalPkg.Lookup(fmt.Sprintf("init.%d", len(target.Inits)))
	regFunc := NewFunc(pos, pos, sym, NewSignature(types.LocalPkg, nil, nil, nil, nil))
	regFunc.Body = body

	target.Inits = append(target.Inits, regFunc)
	// must call this in go1.22
	if needToCreateInit {
		// typecheck.DeclFunc(regFunc)
	} else {
		AddFuncs(regFunc)
	}
}

func takeAddr(fn *ir.Func, field *types.Field, nameOnly bool) ir.Node {
	pos := fn.Pos()
	if field == nil {
		return newNilInterface(pos)
	}
	// go1.20 only? no Nname, so cannot refer to it
	// but we should display it in trace?(better to do so)
	var isNilOrEmptyName bool

	if field.Nname == nil {
		isNilOrEmptyName = true
	} else if field.Sym != nil && field.Sym.Name == "_" {
		isNilOrEmptyName = true
	}
	var fieldRef *ir.Name
	if isNilOrEmptyName {
		// if name is "_" or empty, return nil
		if true {
			return newNilInterface(pos)
		}
		if false {
			// NOTE: not working when IR
			tmpName := typecheck.TempAt(pos, fn, field.Type)
			field.Nname = tmpName
			field.Sym = tmpName.Sym()
			fieldRef = tmpName
		}
	} else {
		fieldRef = field.Nname.(*ir.Name)
	}

	arg := ir.NewAddrExpr(pos, fieldRef)

	if nameOnly {
		fieldType := field.Type
		arg.SetType(types.NewPtr(fieldType))
		return arg
	}

	return convToEFace(pos, arg, field.Type, true)
}

func getFieldName(fn *ir.Func, field *types.Field) string {
	if field == nil {
		return ""
	}
	if field.Nname == nil {
		return ""
	}
	if false {
		// debug
		if field.Sym != nil {
			return field.Sym.Name
		}
	}
	return field.Nname.(*ir.Name).Sym().Name
}

func convToEFace(pos src.XPos, x ir.Node, t *types.Type, ptr bool) *ir.ConvExpr {
	conv := ir.NewConvExpr(pos, ir.OCONV, types.Types[types.TINTER], x)
	conv.SetImplicit(true)

	// go1.20 and above: must have Typeword
	if ptr {
		SetConvTypeWordPtr(conv, t)
	} else {
		SetConvTypeWord(conv, t)
	}
	return conv
}

func isFirstStmtSkipTrap(nodes ir.Nodes) bool {
	// NOTE: for performance reason, only check the first
	if len(nodes) > 0 && isCallTo(nodes[0], xgoRuntimeTrapPkg, "Skip") {
		return true
	}
	if false {
		for _, node := range nodes {
			if isCallTo(node, xgoRuntimeTrapPkg, "Skip") {
				return true
			}
		}
	}
	return false
}

func isCallTo(node ir.Node, pkgPath string, name string) bool {
	callNode, ok := node.(*ir.CallExpr)
	if !ok {
		return false
	}
	nameNode, ok := getCallee(callNode).(*ir.Name)
	if !ok {
		return false
	}
	sym := nameNode.Sym()
	if sym == nil {
		return false
	}
	return sym.Pkg != nil && sym.Name == name && sym.Pkg.Path == pkgPath
}

func newNilInterface(pos src.XPos) ir.Expr {
	return NewNilExpr(pos, types.Types[types.TINTER])
}

func newEmptyStr(pos src.XPos) ir.Node {
	return NewStringLit(pos, "")
}

func newNilInterfaceSlice(pos src.XPos) ir.Expr {
	return NewNilExpr(pos, types.NewSlice(types.Types[types.TINTER]))
}

func getPosInfo(pos src.XPos) src.Pos {
	return base.Ctxt.PosTable.Pos(pos)
}

func getAdjustedFile(f string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	// for windows, posFile has the form:
	//    C:/a/b/c
	// while syncDeclMapping's file has the form:
	//    C:\a\b\c
	return strings.ReplaceAll(f, "/", "\\")
}
