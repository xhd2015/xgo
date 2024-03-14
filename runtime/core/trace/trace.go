package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core/trap"
)

const __XGO_SKIP_TRAP = true

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	panic(errors.New("xgo failed to link __xgo_link_getcurg"))
}

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
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
			stack := &Stack{
				FuncInfo: f,
				Recv:     args.Recv,
				Args:     args.Args,
				Results:  args.Results,
			}
			key := uintptr(__xgo_link_getcurg())
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
			key := uintptr(__xgo_link_getcurg())
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

				// TODO: may add callback for this
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
