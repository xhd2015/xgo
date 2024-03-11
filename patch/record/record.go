package record

import (
	"cmd/compile/internal/ir"
	"fmt"
)

var m map[*ir.Func]ir.Nodes

func GetRewrittenBody(fn *ir.Func) (body ir.Nodes, ok bool) {
	body, ok = m[fn]
	return
}
func SetRewrittenBody(fn *ir.Func, body ir.Nodes) {
	if _, ok := m[fn]; ok {
		panic(fmt.Errorf("rewritten body already set: %v", fn.Nname))
	}
	if m == nil {
		m = make(map[*ir.Func]ir.Nodes)
	}
	m[fn] = body
}

func HasRewritten() bool {
	return len(m) > 0
}
