package main

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
)

// This test verifies if xgo can detect legacy runtime xgo v1.1.0
// and try to avoid variable trapping fallback

func TestMockPatch(t *testing.T) {
	g1 := greet("world")
	if g1 != "hello world" {
		t.Fatalf("expect greet to be hello world, actual: %s", g1)
	}
	mock.Patch(greet, func(name string) string {
		return "patch " + name
	})
	g2 := greet("world")
	if g2 != "patch world" {
		t.Fatalf("expect greet to be patch world, actual: %s", g2)
	}
}
func TestMockInterceptor(t *testing.T) {
	g1 := greet("world")
	if g1 != "hello world" {
		t.Fatalf("expect greet to be hello world, actual: %s", g1)
	}
	mock.Mock(greet, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("mock " + args.GetFieldIndex(0).Value().(string))
		return nil
	})
	g2 := greet("world")
	if g2 != "mock world" {
		t.Fatalf("expect greet to be mock world, actual: %s", g2)
	}
}

func TestTrap(t *testing.T) {
	var records []string
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) (interface{}, error) {
			records = append(records, args.GetFieldIndex(0).Value().(string))
			return nil, nil
		},
	})
	g1 := greet("world")
	if g1 != "hello world" {
		t.Fatalf("expect greet to be hello world, actual: %s", g1)
	}
	if len(records) != 1 {
		t.Fatalf("expect records to be 1, actual: %d", len(records))
	}
	if records[0] != "world" {
		t.Fatalf("expect records to be world, actual: %s", records[0])
	}
}

func TestMockVarShouldSuccess(t *testing.T) {
	before := a
	if before != 10 {
		t.Fatalf("expect a to be 10, actual: %d", before)
	}
	mock.Patch(&a, func() int {
		return 123
	})
	after := a
	if after != 123 {
		t.Fatalf("expect a to be 123, actual: %d", after)
	}
}

func TestAccessVarPtrShouldNotFallbackWhenOnlyVariablePatched(t *testing.T) {
	before := a
	if before != 10 {
		t.Fatalf("expect a to be 10, actual: %d", before)
	}
	mock.Patch(&a, func() int {
		return 123
	})
	ptr := &a
	ptrV := *ptr
	if ptrV != 10 {
		t.Fatalf("expect ptrV to be 10, actual: %d", ptrV)
	}
}
