// This file should be inserted into xgo's runtime to provide hack points
// the 'go:build ignore' line will be removed when copy to goroot

//go:build ignore

package trace_runtime

import (
	"unsafe"
	"runtime"
)

func Runtime_XgoSetTrap(trap func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)) {
	runtime.XgoSetTrap(trap)
}

func Runtime_XgoGetCurG() unsafe.Pointer {
	return runtime.XgoGetCurG()
}

func Runtime_XgoPeekPanic() interface{} {
	return runtime.XgoPeekPanic()
}

func Runtime_XgoGetFullPCName(pc uintptr) string {
	return runtime.XgoGetFullPCName(pc)
}
