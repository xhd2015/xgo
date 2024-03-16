package main

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/functab"

	"github.com/xhd2015/xgo/runtime/pkg"
	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Use()
}

// can break some
func regTest() {

}

func main() {
	// testFindFunc()
	// mock.CheckSSA()
	// runtime.TestModuleData_Requires_Xgo()
	// res := testArgs("a")
	// fmt.Printf("res: %v\n", res)
	A()
	B()
	C()
}

func A() {
	fmt.Printf("A\n")
}

func B() {
	fmt.Printf("B\n")
	C()
}
func C() {
	fmt.Printf("C\n")
}

func testFindFunc() {
	// call registered func
	fn := functab.GetFunc("main.a")
	if fn == nil {
		panic(fmt.Errorf("func main.a not found"))
	}
	fnv := reflect.ValueOf(fn.Func)
	fnv.Call(nil)

	fnb := functab.GetFunc("main.b")
	if fnb == nil {
		panic(fmt.Errorf("func main.b not found"))
	}
	fmt.Printf("main.b recvName=%v,argNames=%v,resNames=%v\n", fnb.RecvName, fnb.ArgNames, fnb.ResNames)
}

// GOSSAFUNC=main.checkSSA ./with-go-devel.sh go build -gcflags="all=-N -l" ./test/test_trap
func checkSSA() {
	var v interface{} = testReflect
	_ = v
	// fmt.Println(testReflect)
}

func getReflectWord(i interface{}) uintptr {
	type IHeader struct {
		typ  uintptr
		word uintptr
	}

	return (*IHeader)(unsafe.Pointer(&i)).word
}

func testReflect() {
	pc := runtime.Getcallerpc()
	entryPC := runtime.GetcallerFuncPC()

	fmt.Printf("testReflect caller pc: %x\n", pc)
	fmt.Printf("testReflect caller entry pc: %x\n", entryPC)
}

func testArgs(s string) int {
	fmt.Printf("testArgs: %s\n", s)

	num(1).add(2)
	return 1
}

type num int

func (c num) add(b int) {
	fmt.Printf("%d+%d=%d\n", c, b, int(c)+b)
	pkg.Hello("pkg")
}

func a() {
	fmt.Printf("main.a called\n")
}
func b(name string) (age int, err error) {
	return
}
