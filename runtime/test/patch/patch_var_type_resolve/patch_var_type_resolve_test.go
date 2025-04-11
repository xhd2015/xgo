package patch_var_type_resolve

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

var a = 10

func TestPatchVarTypeResolveLiteral(t *testing.T) {
	before := a
	if before != 10 {
		t.Fatalf("a should be 10")
	}
	cancel := mock.Patch(&a, func() int {
		return 20
	})

	if a != 20 {
		t.Fatalf("a should be 20")
	}
	cancel()
	after := a
	if after != 10 {
		t.Fatalf("a should be %d", before)
	}
}

type Service struct {
	Hello string
}

func (c *Service) Greet(name string) string {
	return c.Hello + " " + name
}

var varRecv Service = Service{
	Hello: "hello",
}

func demoGetAddr() *Service {
	return &varRecv
}

func TestPatchVarMethodRefPointerAsCaller(t *testing.T) {
	mock.Patch(&varRecv, func() *Service {
		return &Service{
			Hello: "mock",
		}
	})
	if false {
		// this syntax is fine
		demoGetAddr().Greet("world")
	}
	res := varRecv.Greet("world")
	expected := "mock world"
	if res != expected {
		t.Fatalf("expected %s, but got %s", expected, res)
	}
}

func TestPatchVarMethodRefPointerAsArg(t *testing.T) {
	mock.Patch(&varRecv, func() *Service {
		return &Service{
			Hello: "mock",
		}
	})
	if false {
		// this syntax is fine
		demoGetAddr().Greet("world")
	}
	x := varRecv.Greet
	if false {
		t.Logf("%T", x)
	}
	res := x("world")
	expected := "mock world"
	if res != expected {
		t.Fatalf("expected %s, but got %s", expected, res)
	}
}
