package mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/runtime/core/functab"
	"github.com/xhd2015/xgo/runtime/core/trap"
)

// for functions that have trap.AddInterceptor called, trap will skip
// the function and its called function
func Use() {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *functab.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
			if strings.Contains(f.FullName, "testArgs") {
				fmt.Printf("Mock: %s\n", f.FullName)
				p := args.Results[0].(*int)
				*p = 20
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
		Post: func(ctx context.Context, f *functab.FuncInfo, args *trap.FuncArgs, data interface{}) error {
			return nil
		},
	})
}
