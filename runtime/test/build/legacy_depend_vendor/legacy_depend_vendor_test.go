package main

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
)

// This test verifies if xgo can detect legacy runtime version
// and try to upgrade with the one it embeds.
//
// example output:
//
//  xgo test -v ./...
// WARNING: xgo v1.1.0 cannot work with deprecated xgo/runtime v1.0.52.
// xgo will try best to compile with newer xgo/runtime v1.1.0, it's recommended to upgrade via:
//   go get github.com/xhd2015/xgo/runtime@latest
// === RUN   TestMockPatch
// --- PASS: TestMockPatch (0.00s)
// === RUN   TestMockInterceptor
// --- PASS: TestMockInterceptor (0.00s)
// === RUN   TestTrap
// --- PASS: TestTrap (0.00s)
// PASS
// ok      github.com/xhd2015/xgo/testdata/legacy_depend   0.340s

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
			if fn.IdentityName == "greet" {
				records = append(records, args.GetFieldIndex(0).Value().(string))
			}
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
