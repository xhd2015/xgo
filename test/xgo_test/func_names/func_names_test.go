// func names in go has 6 forms:
//    F         plain function
//    func1     closure
//    s.F       method of a struct instance
//    S.F       method of a struct type
//    i.F       method of an interface instance
//    I.F       method of an interface type

// when getting
// when getting functions

package func_names

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func F() {

}

var c = func() {}

type S struct{}

func (S) F() {}
func (S) G() {}

type I interface {
	F()
}

// Embed
type E interface {
	I
	G()
}

type testCase struct {
	fn   interface{}
	name string
}

var getGenericTests func() []*testCase

// go test -run TestFuncNames -v ./test/xgo_test/func_names
// go run ./script/run-test/ --include go1.18.10 --xgo-test-only -run TestFuncNames -v ./test/xgo_test/func_names
func TestFuncNames(t *testing.T) {
	var c3 func()
	c2 := func() {
		c3 = func() {}
	}

	c2()

	var s S
	var i I = s
	var e E = s
	var tests = []*testCase{
		{F, "github.com/xhd2015/xgo/test/xgo_test/func_names.F"},
		{c, expectTopLevelFunc},                                                       // closure
		{c2, "github.com/xhd2015/xgo/test/xgo_test/func_names.TestFuncNames.func1"},   // closure
		{c3, "github.com/xhd2015/xgo/test/xgo_test/func_names.TestFuncNames.func1.1"}, // closure
		{s.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.S.F-fm"},               // -fm suffix
		{S.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.S.F"},
		{i.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.I.F-fm"}, // -fm suffix
		{I.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.I.F"},
		// {Embed.I, ""}, // cannot reference Embed.I: undefined (type Embed has no field or method I)
		{E.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.E.F"},
		{E.G, "github.com/xhd2015/xgo/test/xgo_test/func_names.E.G"},
		{e.F, "github.com/xhd2015/xgo/test/xgo_test/func_names.E.F-fm"},
		{e.G, "github.com/xhd2015/xgo/test/xgo_test/func_names.E.G-fm"},
	}
	if getGenericTests != nil {
		tests = append(tests, getGenericTests()...)
	}
	for _, test := range tests {
		fname := getFuncName(test.fn)
		if fname != test.name {
			t.Fatalf("expect func name: %s, actual: %s", test.name, fname)
		}
	}

	// validate -fm matches
	sfName := getFuncName(s.F)
	SFName := getFuncName(S.F)
	if sfName != SFName+"-fm" {
		t.Fatalf("expect s.F to be:%s-fm, actual: %s", SFName, sfName)
	}

	ifName := getFuncName(i.F)
	IFName := getFuncName(I.F)
	if ifName != IFName+"-fm" {
		t.Fatalf("expect i.F to be:%s-fm, actual: %s", IFName, ifName)
	}
}

func getFuncName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("expect func, actual: %s", v.Kind().String()))
	}
	pc := v.Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		panic(fmt.Errorf("func not found: pc=0x%x", pc))
	}
	return f.Name()
}
