package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/xhd2015/xgo/runtime/core/getg"
	"github.com/xhd2015/xgo/runtime/core/trap"
)

// TODO: remaining problems to be solved:
const __XGO_SKIP_TRAP = true

// hold goroutine stacks, keyed by goroutine ptr
var stackMap sync.Map // uintptr(goroutine) -> *Root

type Root struct {
	// current executed function
	Top      *Stack
	Children []*Stack
}

type Stack struct {
	FuncInfo *trap.FuncInfo
	Recv     interface{}
	Args     []interface{}
	Results  []interface{}
	Children []*Stack
}

func Use() {
	// collect trace
	trap.AddInterceptor(trap.Interceptor{
		Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
			stack := &Stack{
				FuncInfo: f,
				Recv:     args.Recv,
				Args:     args.Args,
				Results:  args.Results,
			}
			key := uintptr(getg.Curg())
			v, ok := stackMap.Load(key)
			if !ok {
				// initial stack
				root := &Root{
					Top: stack,
					Children: []*Stack{
						stack,
					},
				}
				stackMap.Store(key, root)
				return nil, nil
			}
			root := v.(*Root)
			prevTop := root.Top
			root.Top.Children = append(root.Top.Children, stack)
			root.Top = stack
			return prevTop, nil
		},
		Post: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs, data interface{}) error {
			key := uintptr(getg.Curg())
			v, ok := stackMap.Load(key)
			if !ok {
				panic(fmt.Errorf("unbalanced stack"))
			}
			root := v.(*Root)
			if data == nil {
				trace, err := json.Marshal(&Stack{
					Children: root.Children,
				})
				if err != nil {
					return err
				}
				fmt.Printf("trace: %s\n", string(trace))
				stackMap.Delete(key)
			} else {
				// pop stack
				root.Top = data.(*Stack)
			}
			return nil
		},
	})
}
