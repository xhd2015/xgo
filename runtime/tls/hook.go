package tls

import (
	"unsafe"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

func init() {
	runtime.XgoOnCreateG(func(g, childG unsafe.Pointer) {
		mut.Lock()
		localKeys := keys
		mut.Unlock()

		g1 := runtime.AsG(g)
		g2 := runtime.AsG(childG)
		for _, loc := range localKeys {
			if !loc.inherit {
				continue
			}

			v, ok := g1.GetOK(loc)
			if !ok {
				continue
			}
			g2.Set(loc, v)
		}
	})
}
