package interceptor_ctx

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

type Service struct {
}

func (s *Service) Greet(ctx context.Context, name string) (string, error) {
	return "hello " + name, nil
}

func Greet(ctx context.Context, name string) (string, error) {
	return "hello " + name, nil
}

func TestFuncCtx(t *testing.T) {
	var arg string
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "Greet" {
				arg = args.GetFieldIndex(1).Value().(string)
			}
			return nil, nil
		},
	})
	results, err := Greet(context.Background(), "world")
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

func TestTypeFuncReceiverCtx(t *testing.T) {
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

	results, err := svc.Greet(context.Background(), "world")
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

// always expect func receiver ctx
func TestMethodReceiverCtx(t *testing.T) {
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

	results, err := svc.Greet(context.Background(), "world")
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
