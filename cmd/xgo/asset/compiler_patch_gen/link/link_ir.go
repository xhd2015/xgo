package link

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"strings"
)

const xgoPrefix = "__xgo_"

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
		link := findLink(symName[len(xgoPrefix):])
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
