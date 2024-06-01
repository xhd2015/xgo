package func_name

import (
	"cmd/compile/internal/syntax"
	"fmt"
)

func FormatFuncRefName(recvTypeName string, recvPtr bool, funcName string) string {
	if recvTypeName == "" {
		return funcName
	}
	if recvPtr {
		return fmt.Sprintf("(*%s).%s", recvTypeName, funcName)
	}
	return recvTypeName + "." + funcName
}

func FormatFuncRefNameSyntax(pos syntax.Pos, recvTypeName string, recvPtr bool, funcName string) syntax.Expr {
	if recvTypeName == "" {
		return syntax.NewName(pos, funcName)
	}

	var x syntax.Expr
	if recvPtr {
		// return fmt.Sprintf("(*%s).%s", recvTypeName, funcName)
		x = &syntax.ParenExpr{
			X: &syntax.Operation{
				Op: syntax.Mul,
				X:  syntax.NewName(pos, recvTypeName),
			},
		}
	} else {
		x = syntax.NewName(pos, recvTypeName)
	}
	return &syntax.SelectorExpr{
		X:   x,
		Sel: syntax.NewName(pos, funcName),
	}
}
