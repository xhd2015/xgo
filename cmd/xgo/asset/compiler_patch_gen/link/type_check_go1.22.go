//go:build go1.22
// +build go1.22

package link

import (
	"cmd/compile/internal/ir"
)

func typecheckLinkedName(sym *ir.Name) {
	// go1.22 and above does not need to call typecheck
}
