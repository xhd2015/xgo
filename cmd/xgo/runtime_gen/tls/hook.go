package tls

import (
	"fmt"
	"os"
	"unsafe"
)

func __xgo_link_on_gonewproc(f func(g uintptr)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_gonewproc(requires xgo).")
}

func __xgo_link_getcurg() unsafe.Pointer {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_getcurg(requires xgo).")
	return nil
}

func __xgo_link_on_goexit(fn func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_goexit(requires xgo).")
}

func init() {
	__xgo_link_on_goexit(func() {
		// clear when exit
		g := uintptr(__xgo_link_getcurg())
		mut.Lock()
		localKeys := keys
		mut.Unlock()
		for _, loc := range localKeys {
			loc.store.Delete(g)
		}
	})
	__xgo_link_on_gonewproc(func(newg uintptr) {
		// inherit when new goroutine
		g := uintptr(__xgo_link_getcurg())
		mut.Lock()
		localKeys := keys
		mut.Unlock()
		for _, loc := range localKeys {
			if !loc.inherit {
				continue
			}
			val, ok := loc.store.Load(g)
			if !ok {
				continue
			}
			loc.store.Store(newg, val)
		}
	})
}
