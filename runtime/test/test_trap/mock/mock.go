package mock

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core/trap"
)

func Use() {
	trap.AddInterceptor(trap.Interceptor{
		Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
			if strings.Contains(f.FullName, "testArgs") {
				fmt.Printf("Mock: %s\n", f.FullName)
				p := args.Results[0].(*int)
				*p = 20
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
		Post: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs, data interface{}) error {
			return nil
		},
	})
}

type T struct{}

func (c *T) A() {
}

func CheckSSA() {
	println(dumpA)
	println(testReflect)
	type emptyInterface struct {
		typ  uintptr
		word uintptr
	}
	var v interface{} = testReflect

	x := (*emptyInterface)(unsafe.Pointer((&v)))

	typeBase, typeEnd := runtime.TestModuleDataGetType_Requires_Xgo()
	typOff := x.typ - typeBase
	typLen := typeEnd - typeBase

	// see if any offset is typOff

	// 32992
	fmt.Printf("typOff: %d\n", uint(typOff))
	fmt.Printf("typLen: %d\n", uint(typLen))

	idx := runtime.TestModuleDataFindTypeLink_Requires_Xgo(int32(typOff))

	// 416
	fmt.Printf("typOff index: %d\n", idx)

	dumpA(x.typ)

	// fmt.Println(testReflect)
}

// pc=0xa396cea*

// pc=0xa396d4a*
// RAX: 1ab70e0
func dumpA(typ uintptr) {
	println(typ)
}

// var v interface{} = testReflect
// mock.go:28      0x64b9ce0       55              push rbp
// mock.go:28      0x64b9ce1       4889e5          mov rbp, rsp
// mock.go:28      0x64b9ce4*      4883ec18        sub rsp, 0x18
// mock.go:29      0x64b9ce8       488d0539580300  lea rax, ptr [rip+0x35839]
// =>      mock.go:29      0x64b9cef       4889442410      mov qword ptr [rsp+0x10], rax
// mock.go:29      0x64b9cf4       488b442410      mov rax, qword ptr [rsp+0x10]
// mock.go:29      0x64b9cf9       488d0de0a30100  lea rcx, ptr [rip+0x1a3e0]
// mock.go:29      0x64b9d00       48890c24        mov qword ptr [rsp], rcx
// mock.go:29      0x64b9d04       4889442408      mov qword ptr [rsp+0x8], rax
// mock.go:32      0x64b9d09       4883c418        add rsp, 0x18
// mock.go:32      0x64b9d0d       5d              pop rbp
// mock.go:32      0x64b9d0e       c3              ret

// var v interface{} = testReflect
// mock.go:28      0x64b9ce0       55              push rbp
// mock.go:28      0x64b9ce1       4889e5          mov rbp, rsp
// mock.go:28      0x64b9ce4*      4883ec18        sub rsp, 0x18
// mock.go:29      0x64b9ce8       488d0539580300  lea rax, ptr [rip+0x35839]
// =>      mock.go:29      0x64b9cef       4889442410      mov qword ptr [rsp+0x10], rax
// mock.go:29      0x64b9cf4       488b442410      mov rax, qword ptr [rsp+0x10]
// mock.go:29      0x64b9cf9       488d0de0a30100  lea rcx, ptr [rip+0x1a3e0]
// mock.go:29      0x64b9d00       48890c24        mov qword ptr [rsp], rcx
// mock.go:29      0x64b9d04       4889442408      mov qword ptr [rsp+0x8], rax
// mock.go:32      0x64b9d09       4883c418        add rsp, 0x18
// mock.go:32      0x64b9d0d       5d              pop rbp
// mock.go:32      0x64b9d0e       c3              ret

func testReflect() {

}
