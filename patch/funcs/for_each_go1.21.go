//go:build go1.17 && !go1.22
// +build go1.17,!go1.22

package funcs

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
)

// for go1.20, target does not have Target.Funcs, instead, use Target.Decls
func ForEach(callback func(fn *ir.Func) bool) {
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
