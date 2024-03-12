package trap

import (
	"context"
	"errors"
)

const __XGO_SKIP_TRAP = true

var ErrAbort error = errors.New("abort trap interceptor")

type FuncInfo struct {
	Pkg      string
	RecvName string
	RecvPtr  bool
	Name     string

	FullName string
}
type FuncArgs struct {
	Recv    interface{}
	Args    []interface{}
	Results []interface{}
}

type Interceptor struct {
	Pre  func(ctx context.Context, f *FuncInfo, args *FuncArgs) (data interface{}, err error)
	Post func(ctx context.Context, f *FuncInfo, args *FuncArgs, data interface{}) error
}

var interceptors []Interceptor

func AddInterceptor(interceptor Interceptor) {
	interceptors = append(interceptors, interceptor)
}

func GetInterceptors() []Interceptor {
	return interceptors
}
