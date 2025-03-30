package reflect_get_unexported

import (
	"reflect"
	"testing"

	"github.com/xhd2015/xgo/runtime/functab"
)

var mHasBeenCalled bool

// go run ./cmd/xgo test --project-dir runtime -run TestGetUnexportedMethod -v ./test/reflect_get_unexported
func TestGetUnexportedMethod(t *testing.T) {
	s := &struct_{}
	em := reflect.ValueOf(s).MethodByName("M")
	if em.Kind() != reflect.Func {
		t.Fatalf("expect _struct.M to be func, actual: %s", em.Kind().String())
	}
	uem := reflect.ValueOf(s).MethodByName("m")
	if uem.IsValid() {
		t.Fatalf("expect _struct.m through reflect to be invalid, actual: %s", uem.Kind().String())
	}

	st := reflect.TypeOf(s)

	methods := functab.GetTypeMethods(st)
	t.Logf("methods: %v", methods)
	if methods == nil {
		t.Fatalf("expect found methods of struct_, actually not found")
	}

	if methods["m"] == nil {
		t.Fatalf("expect found struct_.m, actually not found")
	}

	if methods["M"] == nil {
		t.Fatalf("expect found struct_.M, actually not found")
	}

	meth := methods["m"]
	if meth == nil {
		t.Fatalf("expect found struct_.m, actually not found")
	}
	mv := reflect.ValueOf(meth.Func)

	// mv.Kind()
	t.Logf("kind: %s", mv.Kind())

	mv.Call([]reflect.Value{reflect.ValueOf(s)})
	if !mHasBeenCalled {
		t.Fatalf("expect found struct_.m has been called, actually not")
	}
}

type struct_ struct {
}

func (c *struct_) M() {}
func (c *struct_) m() {
	if c == nil {
		panic("receiver should be nil")
	}
	mHasBeenCalled = true
}
