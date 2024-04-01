package trap_args

import (
	"context"
	"errors"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func plainErr() error {
	panic("plainErr should be mocked")
}

func subErr() *Error {
	return &Error{"sub error"}
}

func TestPlainErrShouldSetErrRes(t *testing.T) {
	mockErr := errors.New("mock err")
	var err error
	callAndCheck(func() {
		err = plainErr()
	}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
		if !f.LastResultErr {
			t.Fatalf("expect f.LastResultErr to be true, actual: false")
		}
		return mockErr
	})

	if err != mockErr {
		t.Fatalf("expect return err %v, actual %v", mockErr, err)
	}
}

func TestSubErrShouldNotSetErrRes(t *testing.T) {
	mockErr := errors.New("mock err")
	var err *Error
	var recoverErr interface{}
	func() {
		defer func() {
			// NOTE: this may have impact
			trap.Skip()
			recoverErr = recover()
		}()
		callAndCheck(func() {
			err = subErr()
		}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
			if f.LastResultErr {
				t.Fatalf("expect f.LastResultErr to be false, actual: true")
			}
			// even not pl should fail
			return mockErr
		})
	}()

	if recoverErr == nil {
		t.Fatalf("expect error via panic, actually no panic")
	}

	if err != nil {
		t.Fatalf("expect return error not set, actual: %v", err)
	}

	if recoverErr != mockErr {
		t.Fatalf("expect panic err to be %v, actual: %v", mockErr, recoverErr)
	}
}
