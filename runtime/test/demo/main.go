package main

import (
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core/functab"
	"github.com/xhd2015/xgo/runtime/pkg"
)

type T struct {
}

func (c *T) a() {}
func (*T) b()   {}
func (T) c()    {}

type T2 struct{}

func (*T2) b() {}

func reg(v interface{}) {

}
func a() {
	fmt.Printf("main.a called\n")
}
func b(name string) (age int, err error) {
	return
}
func init() {
	v := interface{}(testArgs)
	println(v)

	func() {
		println("closure")
	}()

	reg((*T).a)
	reg((*T).b)
	reg((*T2).b)
	reg(T.c)
	reg(main)
}

func main() {
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
	res := testArgs("a")
	fmt.Printf("res: %v\n", res)
}

func test_DumpIR() {
	after, stop := trap()
	if !stop {
		if after != nil {
			defer after()
		}
		fmt.Printf("hello IR\n")
	}
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

func trap() (after func(), stop bool) {
	return nil, false
}
