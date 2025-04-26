package main

import (
	"os"
	"strconv"
	"unsafe"
)

// check assembly:
//  go build -gcflags="-S" -o /dev/null unsafe_pointer.go

// see https://github.com/xhd2015/xgo/issues/330#issuecomment-2831835483
func main() {
	var ptr uintptr = uintptr(mustParseInt(os.Getenv("TEST")))

	print(123)
	a := unsafePointer(ptr)
	print(456)
	b := unsafePointer2(ptr)
	print(a, b)
}

func unsafePointer(ptr uintptr) unsafe.Pointer {
	return unsafe.Pointer(ptr)
}
func unsafePointer2(ptr uintptr) unsafe.Pointer {
	type P struct {
		ptr unsafe.Pointer
	}
	// p := *(*P)(unsafe.Pointer(&ptr))
	// return p.ptr
	return (*(*P)(unsafe.Pointer(&ptr))).ptr
}

func mustParseInt(v string) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}
	return i
}
