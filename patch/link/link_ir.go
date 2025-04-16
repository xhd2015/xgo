package link

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"strings"
)

const xgoPrefix = "__xgo_"

// LinkXgoInit links function defined in
// runtime
// the function template:
//
//	  var __xgo_trap_0 = func() {}
//
//	  func init() {
//		  __xgo_init_0()
//	  }
//
//	  //go:noinline
//	  func __xgo_init_0() {
//	  	__xgo_trap_0 = __xgo_trap_0
//	  }
func LinkXgoInit(fn *ir.Func) {
	if fn.Body == nil {
		// in go, function can have name without body
		return
	}

	fnName := fn.Sym().Name
	if !strings.HasPrefix(fnName, "__xgo_init_") {
		return
	}
	for _, node := range fn.Body {
		assign, ok := node.(*ir.AssignStmt)
		if !ok {
			continue
		}
		if assign.Def {
			continue
		}
		xName, ok := assign.X.(*ir.Name)
		if !ok {
			continue
		}
		symName := xName.Sym().Name
		if !strings.HasPrefix(symName, xgoPrefix) {
			continue
		}

		switch {
		case strings.HasPrefix(symName, "__xgo_trap_var_"):
			sym := typecheck.LookupRuntime("XgoTrapVar")
			if sym == nil {
				continue
			}
			assign.Y = sym
		case strings.HasPrefix(symName, "__xgo_trap_varptr_"):
			sym := typecheck.LookupRuntime("XgoTrapVarPtr")
			if sym == nil {
				continue
			}
			assign.Y = sym
		case strings.HasPrefix(symName, "__xgo_trap_"):
			sym := typecheck.LookupRuntime("XgoTrap")
			if sym == nil {
				continue
			}
			assign.Y = sym
		case strings.HasPrefix(symName, "__xgo_register_"):
			sym := typecheck.LookupRuntime("XgoRegister")
			if sym == nil {
				continue
			}
			assign.Y = sym
		}
	}
}
