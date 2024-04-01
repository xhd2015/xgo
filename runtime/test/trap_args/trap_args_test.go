package trap_args

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func f1() {

}
func f2(ctx context.Context) error {
	panic("f2 should be mocked")
}

func f3(ctx *MyContext) *Error {
	panic("f3 should be mocked")
}

type MyContext struct {
	context.Context
}

type Error struct {
	msg string
}

func (c *Error) Error() string {
	return c.msg
}

type struct_ struct {
}

func (c *struct_) f2(ctx context.Context) error {
	panic(fmt.Errorf("struct_.f2 should be mocked"))
}
func TestCtxArgWithRecvCanBeRecognized(t *testing.T) {
	st := &struct_{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test", "mock")

	var callCount int
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			callCount++
			if !f.FirstArgCtx {
				t.Fatalf("expect first arg to be context")
			}
			if trapCtx != ctx {
				t.Fatalf("expect context passed unchanged, actually different")
			}
			ctxVal := trapCtx.Value("test").(string)
			if ctxVal != "mock" {
				t.Fatalf("expect context value to be %q, actual: %q", "mock", ctxVal)
			}
			return nil, trap.ErrAbort
		},
	})
	st.f2(ctx)

	if callCount != 1 {
		t.Fatalf("expect call trap once, actual: %d", callCount)
	}
}

func callAndCheck(fn func(), check func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error) error {
	pc := reflect.ValueOf(fn).Pointer()

	pcName := runtime.FuncForPC(pc).Name()
	if pcName == "" {
		return fmt.Errorf("cannot get pc name")
	}

	var callCount int
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if callCount == 0 {
				callCount++
				if pcName != f.FullName {
					panic(fmt.Errorf("expect first call hit %s, actual: %s", pcName, f.FullName))
				}
				return
			}
			callCount++
			err = check(trapCtx, f, args, result)
			if err != nil {
				return nil, err
			}
			return nil, trap.ErrAbort
		},
	})
	fn()

	if callCount != 2 {
		fmt.Errorf("expect call trap twice, actual: %d", callCount)
	}
	return nil
}
