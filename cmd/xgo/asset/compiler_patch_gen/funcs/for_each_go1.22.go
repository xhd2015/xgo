//go:build go1.22
// +build go1.22

package funcs

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
)

func ForEach(callback func(fn *ir.Func) bool) {
	for _, fn := range typecheck.Target.Funcs {
		if !callback(fn) {
			return
		}
	}
}
