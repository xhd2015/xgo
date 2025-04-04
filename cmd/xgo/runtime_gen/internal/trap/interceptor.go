package trap

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core"
)

// concept:
//   - mock: associated with a function or variable that is meant to replace the original function
//   - recorder: associated with a specific function or variable that is meant to record the function call and results
//   - interceptor: general purpose interceptor, not associated with a specific function or variable, can be used to mock, record, etc.
//
// chain behavior:
//    - mock cannot be chained, only the last one will take effect
//    - interceptors and recorders can be chained

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

type PreInterceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) (interface{}, error)

type PostInterceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object, data interface{}) error

func PushInterceptor(pre PreInterceptor, post PostInterceptor) func() {
	return pushInterceptor(pre, post)
}

func PushRecorderInterceptor(fn interface{}, preInterceptor PreInterceptor, postInterceptor PostInterceptor) func() {
	return pushRecorderInterceptor(fn, preInterceptor, postInterceptor)
}

func pushInterceptor(pre PreInterceptor, post PostInterceptor) func() {
	preHandler, postHandler := buildRecorderFromInterceptor(nil, pre, post)
	stack := getOrAttachStackData()
	if stack.interceptors == nil {
		stack.interceptors = []*recorderHolder{}
	}
	h := &recorderHolder{pre: preHandler, post: postHandler}
	stack.interceptors = append(stack.interceptors, h)
	return func() {
		list := stack.interceptors
		n := len(list)
		if list[n-1] == h {
			stack.interceptors = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				stack.interceptors = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop interceptor not found, check if the interceptor is already popped earlier"))
	}
}
