package name_start_withtest

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

type Test struct {
	Hello string
}

var TestVar = Test{Hello: "hello"}

func TestNameStartWithTest(t *testing.T) {
	res := TestVar.Hello
	if res != "hello" {
		t.Errorf("expect TestVar.Hello = %v, but got %v", "hello", res)
	}
	mock.Patch(&TestVar, func() Test {
		return Test{Hello: "mock"}
	})
	result := TestVar.Hello
	expected := "mock"
	if result != expected {
		t.Errorf("expect TestVar.Hello = %v, but got %v", expected, result)
	}
}
