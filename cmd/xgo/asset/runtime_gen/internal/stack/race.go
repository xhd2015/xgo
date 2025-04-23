//go:build race
// +build race

package stack

import "unsafe"

// see https://github.com/xhd2015/xgo/issues/330
func unsafePointer(ptr uintptr) unsafe.Pointer {
	type P struct {
		ptr unsafe.Pointer
	}
	p := (*P)(unsafe.Pointer(&ptr))

	return p.ptr
}
