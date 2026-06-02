package link

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"fmt"
	"os"
	"path/filepath"
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
		From: "trap_var_",
		To:   "XgoTrapVar",
	},
	{
		From: "trap_varptr_",
		To:   "XgoTrapVarPtr",
	},
	{
		From: "trap_",
		To:   "XgoTrap",
	},
	{
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

func LinkXgoInit(fn *ir.Func) {
	if fn.Body == nil {
		return
	}

	fnName := fn.Sym().Name
	if !strings.HasPrefix(fnName, "__xgo_init_") {
		return
	}
	appendDebug("LINK_INIT fn=%s body_stmts=%d", fnName, len(fn.Body))
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
		appendDebug("LINK_STMT sym=%s link=%s", symName, link)
		sym := typecheck.LookupRuntime(link)
		if sym == nil {
			appendDebug("LINK_STMT LookupRuntime(%s) returned NIL", link)
			continue
		}
		typecheckLinkedName(sym)
		assign.Y = sym
	}
}

func appendDebug(format string, args ...interface{}) {
	logPath := filepath.Join(os.TempDir(), "xgo_link_init_debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "XGO_LOG_ERROR: cannot open %s: %v\n", logPath, err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, format+"\n", args...)
}
