package interceptor_ctx

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestFuncCustomCtx(t *testing.T) {
	var arg string
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "Greet" {
				arg = args.GetFieldIndex(1).Value().(string)
			}
			return nil, nil
		},
	})
	ctx := NewCustomContext()
	results, err := Greet(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "hello world" {
		t.Errorf("expect hello world, but got %s", results)
	}
	if arg != "world" {
		t.Errorf("expect world, but got %s", arg)
	}
}

func TestTypeFuncReceiverCustomCtx(t *testing.T) {
	svc := &Service{}
	var arg string
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "(*Service).Greet" {
				// 0: recv
				// 1: ctx
				// 2: name
				arg = args.GetFieldIndex(2).Value().(string)
			}
			return nil, nil
		},
	})

	ctx := NewCustomContext()
	results, err := svc.Greet(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "hello world" {
		t.Errorf("expect hello world, but got %s", results)
	}
	if arg != "world" {
		t.Errorf("expect world, but got %s", arg)
	}
}

func TestPostFuncCustomCtx(t *testing.T) {
	var postCtx context.Context
	var postName string
	trap.AddInterceptor(&trap.Interceptor{
		Post: func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object, data interface{}) error {
			if f.IdentityName == "GreetCustomCtx" {
				postCtx = ctx
				postName = args.GetFieldIndex(1).Value().(string)
			}
			return nil
		},
	})
	ctx := NewCustomContext()
	results, err := GreetCustomCtx(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "hello world" {
		t.Errorf("expect hello world, but got %s", results)
	}
	if postName != "world" {
		t.Errorf("expect world, but got %s", postName)
	}
	if postCtx == nil {
		t.Error("expect post ctx to be set")
	}
}

func TestPostTypeFuncReceiverCustomCtx(t *testing.T) {
	svc := &Service{}
	var postCtx context.Context
	var postName string
	trap.AddInterceptor(&trap.Interceptor{
		Post: func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object, data interface{}) error {
			if f.IdentityName == "(*Service).Greet" {
				postCtx = ctx
				postName = args.GetFieldIndex(2).Value().(string)
			}
			return nil
		},
	})
	ctx := NewCustomContext()
	results, err := svc.Greet(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "hello world" {
		t.Errorf("expect hello world, but got %s", results)
	}
	if postName != "world" {
		t.Errorf("expect world, but got %s", postName)
	}
	if postCtx == nil {
		t.Error("expect post ctx to be set")
	}
}
