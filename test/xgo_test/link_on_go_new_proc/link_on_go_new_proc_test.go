package trap_set

import (
	"testing"
	"unsafe"
)

func __xgo_link_on_gonewproc(f func(g uintptr)) {
	panic("failed to link __xgo_link_on_gonewproc(requires xgo).")
}

func __xgo_link_getcurg() unsafe.Pointer {
	panic("failed to link __xgo_link_getcurg(requires xgo).")
}

func TestLinkOnGoNewPrco(t *testing.T) {
	var newg uintptr
	__xgo_link_on_gonewproc(func(g uintptr) {
		newg = g
	})

	var rang uintptr
	done := make(chan struct{})
	go func() {
		rang = uintptr(__xgo_link_getcurg())
		close(done)
	}()

	<-done

	if newg == 0 {
		t.Fatalf("expect newg captured, actually not")
	}
	if newg != rang {
		t.Fatalf("expect newg to be 0x%x, actual: 0x%x", rang, newg)
	}

}
