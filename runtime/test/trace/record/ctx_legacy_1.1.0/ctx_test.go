package ctx

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
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
	trace.RecordCall(Greet, func(ctx context.Context, name string, res string, err error) {
		arg = name
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

	myCtx := context.WithValue(context.Background(), "myKey", "myValue")
	trace.RecordCall((*Service).Greet, func(recv *Service, ctx context.Context, name string, res string, err error) {
		if ctx != myCtx {
			t.Errorf("expect myCtx, but got %v", ctx)
		}
		arg = name
	})

	results, err := svc.Greet(myCtx, "world")
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

	myCtx := context.WithValue(context.Background(), "myKey", "myValue")
	trace.RecordCall(svc.Greet, func(ctx context.Context, name string, res string, err error) {
		if ctx != myCtx {
			t.Errorf("expect myCtx, but got %v", ctx)
		}
		arg = name
	})

	results, err := svc.Greet(myCtx, "world")
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
