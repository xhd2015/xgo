package simple

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func greet(s string) string {
	return "hello " + s
}

func TestGreet(t *testing.T) {
	res := greet("world")
	if res != "hello world" {
		t.Errorf("expect greet() to be %s, but got %s", "hello world", res)
	}
	mock.Patch(greet, func(s string) string {
		return "mock " + s
	})
	res = greet("world")
	if res != "mock world" {
		t.Errorf("expect greet() to be %s, but got %s", "mock world", res)
	}
}
