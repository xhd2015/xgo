package main

import (
	"fmt"
	"reflect"

	"golang.org/x/sys/unix"
)

func main() {
	v := unix.Syscall
	fmt.Printf("syscall: 0x%x\n", reflect.ValueOf(v).Pointer())
}
