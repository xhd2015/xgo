package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
	"golang.org/x/net/netutil"
)

func greet(name string) string {
	return "hello " + name
}

func main() {
	fmt.Println(greet("world"))

	if os.Getenv("TEST_MOCK_PATCH") == "true" {
		t := testing.T{}
		exampleMockPatch(&t)
		exampleMockInterceptor(&t)
		exampleTrap(&t)

		netutil.LimitListener(nil, 1)
	}

}

func exampleMockPatch(t *testing.T) {
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
func exampleMockInterceptor(t *testing.T) {
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

func exampleTrap(t *testing.T) {
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
