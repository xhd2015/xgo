package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_var/third"
)

var svcLiteral = third.GreetService{}
var svcFunc = third.NewService()

func TestLiteralVarMethod(t *testing.T) {
	mock.Patch(svcLiteral.Greet, func(name string) string {
		return "mock " + name
	})
	result := svcLiteral.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.Greet() = %v, but got %v", expected, result)
	}
}

// func TestFuncMethod(t *testing.T) {
// 	mock.Patch(svcFunc.Greet, func(name string) string {
// 		return "mock " + name
// 	})
// 	result := svcFunc.Greet("world")
// 	expected := "mock world"
// 	if result != expected {
// 		t.Errorf("expect svc.Greet() = %v, but got %v", expected, result)
// 	}
// }
