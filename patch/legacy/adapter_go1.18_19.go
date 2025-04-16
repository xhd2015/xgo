//go:build go1.18 && !go1.20
// +build go1.18,!go1.20

package patch

import (
	"cmd/compile/internal/ir"
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	"strings"
)

// only need for go1.18
func isClosureWrapperForGeneric(fn *ir.Func) bool {
	if !xgo_ctxt.GenericImplIsClosure {
		return false
	}
	body := fn.Body
	if len(body) != 1 {
		return false
	}
	stmt := body[0]
	var callExpr *ir.CallExpr
	if rt, ok := stmt.(*ir.ReturnStmt); ok {
		if len(rt.Results) != 1 {
			return false
		}
		resCall, ok := rt.Results[0].(*ir.CallExpr)
		if !ok {
			return false
		}
		callExpr = resCall
	} else if cl, ok := stmt.(*ir.CallExpr); ok {
		callExpr = cl
	}
	if callExpr == nil {
		return false
	}
	name, _ := callExpr.X.(*ir.Name)
	if name == nil || name.Sym() == nil || !strings.Contains(name.Sym().Name, "[") {
		return false
	}
	// nArgs=3: arg[1] = receiver
	nArgs := len(callExpr.Args)
	if nArgs != 2 && nArgs != 3 {
		return false
	}
	addr, _ := callExpr.Args[0].(*ir.AddrExpr)
	if addr == nil {
		return false
	}
	addrName, _ := addr.X.(*ir.Name)
	if addrName == nil || addrName.Sym() == nil || !strings.HasPrefix(addrName.Sym().Name, ".dict") {
		return false
	}

	return true
}
