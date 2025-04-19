package main

import (
	"fmt"
	"unsafe"
)

func main() {
	res := nosplitSmall(nil)
	fmt.Printf("hello nosplit:%v\n", res)
}

//go:nosplit
func nosplitSmall(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
