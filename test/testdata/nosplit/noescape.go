package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Printf("hello nosplit:%v\n", noescape(nil))
}

//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
