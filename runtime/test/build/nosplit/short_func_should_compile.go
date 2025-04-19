package nosplit

import (
	"unsafe"
)

//go:nosplit
func nosplitSmall(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
