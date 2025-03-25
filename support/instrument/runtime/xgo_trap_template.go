// This file should be inserted into runtime to provide hack points
// the 'go:build ignore' line will be removed when copy to goroot

//go:build ignore

package runtime

import (
	"unsafe"
)

// make unsafe always imported
var _ = unsafe.Pointer(nil)

// this signature is to avoid relying on defined type
// just raw string,interface{},[]string, []interface{} are needed
var __xgo_trap func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)

func XgoTrap(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool) {
	if __xgo_trap == nil {
		return nil, false
	}
	return __xgo_trap(recvName, recvPtr, argNames, args, resultNames, results)
}

func XgoSetTrap(trap func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)) {
	if __xgo_trap != nil {
		panic("__xgo_trap already set")
	}
	__xgo_trap = trap
}

// __xgo_g is a wrapper around the runtime.G struct
// to avoid exposing the runtime.G struct to the user
// and to avoid having to import the runtime package
// in the user's code.
type __xgo_g struct {
	goid       int64
	parentGoID int64

	gls                 map[interface{}]interface{}
	looseJsonMarshaling bool
}

func XgoGetCurG() unsafe.Pointer {
	curg := getg().m.curg
	if curg.__xgo_g == nil {
		curg.__xgo_g = &__xgo_g{
			goid: curg.goid,
			// parentGoID: curg.parentGoid,
			gls: make(map[interface{}]interface{}),
		}
	}
	return unsafe.Pointer(curg.__xgo_g)
}

// Peek panic without recover
// check gorecover() for implementation details
func XgoPeekPanic() interface{} {
	gp := getg()
	p := gp._panic
	if p == nil || p.goexit || p.recovered {
		return nil
	}
	return p.arg
}

func XgoIsLooseJsonMarshaling() bool {
	curg := getg().m.curg
	return curg.__xgo_g != nil && curg.__xgo_g.looseJsonMarshaling
}

// XgoGetFullPCName returns full name
// without ellipsis
func XgoGetFullPCName(pc uintptr) string {
	return __xgo_get_pc_name_impl(pc)
}
