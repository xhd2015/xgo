package err_mocked

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
	"golang.org/x/example/hello/reverse"
)

func TestNotMocked(t *testing.T) {
	var trapped bool
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Pkg == "golang.org/x/example/hello/reverse" && f.IdentityName == "String" {
				trapped = true
				result.GetFieldIndex(0).Set("mocked")
				return nil, nil
			}
			return nil, nil
		},
	})
	res := reverse.String("hello")
	expect := "olleh"
	if res != expect {
		t.Errorf("expect %s, actual %s", expect, res)
	}
	if !trapped {
		t.Errorf("expect trapped, actually not trapped")
	}
}

func TestMocked(t *testing.T) {
	var trapped bool
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Pkg == "golang.org/x/example/hello/reverse" && f.IdentityName == "String" {
				trapped = true
				result.GetFieldIndex(0).Set("mocked")
				return nil, trap.ErrMocked
			}
			return nil, nil
		},
	})
	res := reverse.String("hello")
	expect := "mocked"
	if res != expect {
		t.Errorf("expect %s, actual %s", expect, res)
	}
	if !trapped {
		t.Errorf("expect trapped, actually not trapped")
	}
}
