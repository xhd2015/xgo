//go:build !go1.22
// +build !go1.22

package link

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
)

func typecheckLinkedName(sym *ir.Name) {
	// go1.21 need to explicitly call typecheck
	typecheck.Expr(sym)
}
