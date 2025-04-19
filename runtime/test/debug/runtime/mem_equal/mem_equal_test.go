package mem_equal

import (
	"testing"
	"unsafe"
)

func __xgo_link_mem_equal(a, b unsafe.Pointer, size uintptr) bool {
	return false
}

// xgo test -run TestMemEqual -v ./test/mem_equal
func TestMemEqual(t *testing.T) {
	t.Skip("Maybe supported later")
	const N = 512
	var a [N]int
	var b [N]int
	var c [N]int

	c[10] = 1

	if !__xgo_link_mem_equal(unsafe.Pointer(&a), unsafe.Pointer(&b), N) {
		t.Fatalf("expect mem_equal(a,b) == true, actual: false")
	}

	if __xgo_link_mem_equal(unsafe.Pointer(&b), unsafe.Pointer(&c), N) {
		t.Fatalf("expect mem_equal(a,b) == false, actual: true")
	}
}
