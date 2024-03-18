package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

const __XGO_SKIP_TRAP = true

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	panic(errors.New("failed to link __xgo_link_getcurg"))
}

// hold goroutine stacks, keyed by goroutine ptr
var stackMap sync.Map // uintptr(goroutine) -> *Root

type Root struct {
	// current executed function
	Top      *Stack
	Children []*Stack
}

type Stack struct {
	FuncInfo *core.FuncInfo

	Args    core.Object
	Results core.Object
	// Recv     interface{}
	// Args     []interface{}
	// Results  []interface{}
	Children []*Stack
}

func Use() {
	// collect trace
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object) (interface{}, error) {
			trap.Skip()
			stack := &Stack{
				FuncInfo: f,
				Args:     args,
				Results:  results,
				// Recv:     args.Recv,
				// Args:     args.Args,
				// Results:  args.Results,
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
		Post: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object, data interface{}) error {
			trap.Skip()
			key := uintptr(__xgo_link_getcurg())
			v, ok := stackMap.Load(key)
			if !ok {
				panic(fmt.Errorf("unbalanced stack"))
			}
			root := v.(*Root)
			if data == nil {
				// stack finished
				stackMap.Delete(key)
				err := emitTrace(&Stack{
					Children: root.Children,
				})
				if err != nil {
					return err
				}
			} else {
				// pop stack
				root.Top = data.(*Stack)
			}
			return nil
		},
	})
}

// this should also be marked as trap.Skip()
func emitTrace(stack *Stack) error {
	// write to file
	trace, err := json.Marshal(stack)
	if err != nil {
		return err
	}

	traceIDNum := int64(1)
	ghex := fmt.Sprintf("g_%x", __xgo_link_getcurg())
	traceID := "t_" + strconv.FormatInt(traceIDNum, 10)

	xgoTraceOutput := os.Getenv("XGO_TRACE_OUTPUT")
	if xgoTraceOutput == "" {
		xgoTraceOutput = time.Now().Format("trace_20060102_150405")
	}
	if xgoTraceOutput == "stdout" {
		// TODO: may add callback for this
		fmt.Printf("%s/%s: ", ghex, traceID)
		fmt.Println(string(trace))
		return nil
	}

	dir := filepath.Join(xgoTraceOutput, ghex)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	file := filepath.Join(dir, traceID+".json")

	return ioutil.WriteFile(file, trace, 0755)
}
