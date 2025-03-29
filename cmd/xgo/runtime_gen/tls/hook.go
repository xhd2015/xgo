package tls

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/legacy"
)

func __xgo_link_on_gonewproc(f func(g uintptr)) {
	if !legacy.V1_0_0 {
		return
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_gonewproc(requires xgo).")
}

func __xgo_link_getcurg() unsafe.Pointer {
	if !legacy.V1_0_0 {
		return nil
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_getcurg(requires xgo).")
	return nil
}

func __xgo_link_on_goexit(fn func()) {
	if !legacy.V1_0_0 {
		return
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_goexit(requires xgo).")
}

func __xgo_link_is_system_stack() bool {
	if !legacy.V1_0_0 {
		return false
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_is_system_stack(requires xgo).")
	return false
}

func init() {
	__xgo_link_on_goexit(func() {
		// clear when exit
		g := uintptr(__xgo_link_getcurg())

		// see https://github.com/xhd2015/xgo/issues/96
		sysStack := __xgo_link_is_system_stack()
		if !sysStack {
			mut.Lock()
		}

		localKeys := keys
		if !sysStack {
			mut.Unlock()
		}

		// NOTE: we must delete entries, otherwise they would
		// cause memory leak
		for _, loc := range localKeys {
			loc.store.Delete(g)
		}
	})

	__xgo_link_on_gonewproc(func(newg uintptr) {
		// cannot lock/unlock on sys stack
		if __xgo_link_is_system_stack() {
			return
		}

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
