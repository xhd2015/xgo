package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/third"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/third/embed_other_struct"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/third/intf"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/third/struct_field"
)

func TestGreet(t *testing.T) {
	mock.Patch(third.Greet, func(name string) string {
		return "mock " + name
	})
	result := third.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect Greet() = %v, but got %v", expected, result)
	}
}

func TestStructGreet(t *testing.T) {
	svc := third.GreetService{}
	mock.Patch(svc.Greet, func(name string) string {
		return "mock " + name
	})
	result := svc.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.Greet() = %v, but got %v", expected, result)
	}
}

func TestInterfaceMethodShouldPanic(t *testing.T) {
	isvc := intf.NewGreetService()

	var panicErr interface{}
	// expect panic
	func() {
		defer func() {
			panicErr = recover()
		}()
		mock.Patch(isvc.Greet, func(name string) string {
			return "mock " + name
		})
		isvc.Greet("world")
	}()
	if panicErr == nil {
		t.Errorf("expect panic")
	}
}

func TestStructValueField(t *testing.T) {
	svc := struct_field.SomeStruct{}
	mock.Patch(svc.MyField.Greet, func(name string) string {
		return "mock " + name
	})
	result := svc.MyField.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.MyField.Greet() = %v, but got %v", expected, result)
	}
}

func TestStructPtrField(t *testing.T) {
	svc := &struct_field.SomeStruct{}
	mock.Patch(svc.MyField.Greet, func(name string) string {
		return "mock " + name
	})
	result := svc.MyField.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.MyField.Greet() = %v, but got %v", expected, result)
	}
}

func TestEmbedOtherStructValue(t *testing.T) {
	svc := embed_other_struct.EmbedStruct{}
	mock.Patch(svc.Other.Greet, func(name string) string {
		return "mock " + name
	})
	result := svc.Other.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.Other.Greet() = %v, but got %v", expected, result)
	}
}

func TestEmbedOtherStructPtr(t *testing.T) {
	svc := &embed_other_struct.EmbedStruct{}
	mock.Patch(svc.Other.Greet, func(name string) string {
		return "mock " + name
	})
	result := svc.Other.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect svc.Other.Greet() = %v, but got %v", expected, result)
	}
}
