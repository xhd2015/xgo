package interceptor_ctx

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
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
	mock.Mock(Greet, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("mock " + args.GetFieldIndex(1).Value().(string))
		return nil
	})

	results, err := Greet(context.Background(), "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "mock world" {
		t.Fatal("results not equal")
	}
}

func TestTypeFuncReceiverCtx(t *testing.T) {
	svc := &Service{}
	mock.Mock((*Service).Greet, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("mock " + args.GetFieldIndex(2).Value().(string))
		return nil
	})

	results, err := svc.Greet(context.Background(), "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "mock world" {
		t.Fatal("results not equal")
	}
}

func TestMethodReceiverCtx(t *testing.T) {
	svc := &Service{}
	mock.Mock(svc.Greet, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("mock " + args.GetFieldIndex(1).Value().(string))
		return nil
	})

	results, err := svc.Greet(context.Background(), "world")
	if err != nil {
		t.Fatal(err)
	}
	if results != "mock world" {
		t.Fatal("results not equal")
	}
}
