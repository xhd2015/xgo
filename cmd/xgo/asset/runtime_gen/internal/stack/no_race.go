//go:build !race
// +build !race

package stack

import (
	"unsafe"
)

func unsafePointer(ptr uintptr) unsafe.Pointer {
	return unsafe.Pointer(ptr)
}
