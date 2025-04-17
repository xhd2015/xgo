package link

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"strings"
)

const XGO_PREFIX = "__xgo_"
const LEGACY_PREFIX = "runtime_link_"

type Link struct {
	From string
	To   string
}

var links = []*Link{
	{
		// From: "__xgo_trap_var_",
		From: "trap_var_",
		To:   "XgoTrapVar",
	},
	{
		// From: "__xgo_trap_varptr_",
		From: "trap_varptr_",
		To:   "XgoTrapVarPtr",
	},
	{
		// From: "__xgo_trap_",
		From: "trap_",
		To:   "XgoTrap",
	},
	{
		// From: "__xgo_register_",
		From: "register_",
		To:   "XgoRegister",
	},
}

func findLink(name string) string {
	for _, link := range links {
		if strings.HasPrefix(name, link.From) {
			return link.To
		}
	}
	return ""
}

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
	linkStmts(fn.Body, fnName == "__xgo_init_legacy_v1_1_0")
}

func linkStmts(stmts ir.Nodes, legacy bool) {
	for _, node := range stmts {
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
		var link string
		if !legacy {
			if !strings.HasPrefix(symName, XGO_PREFIX) {
				continue
			}
			link = findLink(symName[len(XGO_PREFIX):])
		} else {
			if !strings.HasPrefix(symName, LEGACY_PREFIX) {
				continue
			}
			link = symName[len(LEGACY_PREFIX):]
		}
		if link == "" {
			continue
		}
		sym := typecheck.LookupRuntime(link)
		if sym == nil {
			continue
		}
		typecheckLinkedName(sym)
		assign.Y = sym
	}
}
