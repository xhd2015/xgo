package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party/third/struct_field"
)

func TestStructField(t *testing.T) {
	t.Skip("only for debug")
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
