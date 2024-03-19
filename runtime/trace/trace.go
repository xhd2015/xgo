package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

const __XGO_SKIP_TRAP = true

// hold goroutine stacks, keyed by goroutine ptr
var stackMap sync.Map       // uintptr(goroutine) -> *Root
var testInfoMaping sync.Map // uintptr(goroutine) -> *testInfo

type testInfo struct {
	name string
}

func init() {
	__xgo_link_on_test_start(func(t *testing.T, fn func(t *testing.T)) {
		name := t.Name()
		if name == "" {
			return
		}
		key := uintptr(__xgo_link_getcurg())
		testInfoMaping.LoadOrStore(key, &testInfo{
			name: name,
		})
	})
	__xgo_link_on_goexit(func() {
		key := uintptr(__xgo_link_getcurg())
		testInfoMaping.Delete(key)
	})
}

// link by compiler
func __xgo_link_on_test_start(fn func(t *testing.T, fn func(t *testing.T))) {
}

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	panic(errors.New("failed to link __xgo_link_getcurg"))
}

func __xgo_link_on_goexit(fn func()) {
	panic("failed to link __xgo_link_on_goexit")
}

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
	if getTraceOutput() == "off" {
		return
	}
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

func getTraceOutput() string {
	return os.Getenv("XGO_TRACE_OUTPUT")
}

var marshalStack func(stack *Stack) ([]byte, error)

func SetMarshalStack(fn func(stack *Stack) ([]byte, error)) {
	marshalStack = fn
}

func fmtStack(stack *Stack) (data []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
			return
		}
	}()
	if marshalStack != nil {
		return marshalStack(stack)
	}
	return json.Marshal(stack)
}

// this should also be marked as trap.Skip()
// TODO: may add callback for this
func emitTrace(stack *Stack) error {
	var testName string

	key := uintptr(__xgo_link_getcurg())
	tinfo, ok := testInfoMaping.Load(key)
	if ok {
		testName = tinfo.(*testInfo).name
	}

	xgoTraceOutput := getTraceOutput()
	useStdout := xgoTraceOutput == "stdout"
	subName := testName
	if testName == "" {
		traceIDNum := int64(1)
		ghex := fmt.Sprintf("g_%x", __xgo_link_getcurg())
		traceID := "t_" + strconv.FormatInt(traceIDNum, 10)
		if xgoTraceOutput == "" {
			traceDir := time.Now().Format("trace_20060102_150405")
			subName = filepath.Join(traceDir, ghex, traceID)
		} else if useStdout {
			subName = fmt.Sprintf("%s/%s", ghex, traceID)
		} else {
			subName = filepath.Join(xgoTraceOutput, ghex, traceID)
		}
	}

	if useStdout {
		fmt.Printf("%s: ", subName)
	}
	var traceOut []byte
	trace, stackErr := fmtStack(stack)
	if stackErr != nil {
		traceOut = []byte("error:" + stackErr.Error())
	} else {
		traceOut = trace
	}

	if useStdout {
		fmt.Print(traceOut)
		return nil
	}

	subFile := subName + ".json"
	subDir := filepath.Dir(subFile)
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(subFile, traceOut, 0755)
}
