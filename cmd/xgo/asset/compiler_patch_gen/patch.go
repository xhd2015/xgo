package patch

import (
	"cmd/compile/internal/ir"

	"cmd/compile/internal/xgo_rewrite_internal/patch/funcs"
	"cmd/compile/internal/xgo_rewrite_internal/patch/link"
)

func Patch() {
	linkFuncs()
}

func linkFuncs() {
	funcs.ForEach(func(fn *ir.Func) bool {
		link.LinkXgoInit(fn)
		return true
	})
}
