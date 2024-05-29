package trap

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestTrapReturnedErrorShouldBeSet(t *testing.T) {
	mockErr := errors.New("must not be an odd")
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			return nil, mockErr
		},
	})
	err := mustBeOdd(10)
	if err == nil {
		t.Fatalf("expect mustBeOdd(10) returns err, actually no error")
	}
	if err != mockErr {
		t.Fatalf("expect mustBeOdd(10)'s error to be mocked as %v, actual: %v", mockErr, err)
	}
}

func TestTrapPosReturnedErrorShouldBeSet(t *testing.T) {
	mockErr := errors.New("must not be an odd")
	trap.AddInterceptor(&trap.Interceptor{
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) error {
			if f.Stdlib {
				return nil
			}
			return mockErr
		},
	})
	err := mustBeOdd(10)
	if err == nil {
		t.Fatalf("expect mustBeOdd(10) returns err, actually no error")
	}
	if err != mockErr {
		t.Fatalf("expect mustBeOdd(10)'s error to be mocked as %v, actual: %v", mockErr, err)
	}
}

func mustBeOdd(i int) error {
	if i%2 == 0 {
		return fmt.Errorf("not an odd")
	}
	return nil
}
