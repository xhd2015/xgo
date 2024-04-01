package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

const __XGO_SKIP_TRAP = true

// hold goroutine stacks, keyed by goroutine ptr
var stackMap sync.Map       // uintptr(goroutine) -> *Root
var testInfoMapping sync.Map // uintptr(goroutine) -> *testInfo

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
		testInfoMapping.LoadOrStore(key, &testInfo{
			name: name,
		})
	})
	__xgo_link_on_goexit(func() {
		key := uintptr(__xgo_link_getcurg())
		testInfoMapping.Delete(key)
		collectingMap.Delete(key)
	})
}

// link by compiler
func __xgo_link_on_test_start(fn func(t *testing.T, fn func(t *testing.T))) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_test_start(requires xgo).")
}

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_getcurg(requires xgo).")
	return nil
}

func __xgo_link_on_goexit(fn func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_goexit(requires xgo).")
}
func __xgo_link_init_finished() bool {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_init_finished(requires xgo).")
	return false
}

// linked by compiler
func __xgo_link_peek_panic() interface{} {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_peek_panic(requires xgo).")
	return nil
}

var enabledGlobal int32

func Enable() {
	if getTraceOutput() == "off" {
		return
	}
	if __xgo_link_init_finished() {
		panic("Enable must be called from init")
	}
	if !atomic.CompareAndSwapInt32(&enabledGlobal, 0, 1) {
		return
	}
	setupInterceptor()
}

// executes f and collect its trace
// by default trace output will be
// controlled by XGO_TRACE_OUTPUT
func Collect(f func()) {
	if !__xgo_link_init_finished() {
		panic("Collect cannot be called from init")
	}
	collect(f, &collectOpts{})
}

type collectOpts struct {
	name       string
	onComplete func(root *Root)
	root       *Root
}

func Options() *collectOpts {
	return &collectOpts{}
}

func (c *collectOpts) Name(name string) *collectOpts {
	c.name = name
	return c
}

func (c *collectOpts) OnComplete(f func(root *Root)) *collectOpts {
	c.onComplete = f
	return c
}

func (c *collectOpts) Collect(f func()) {
	collect(f, c)
}

func setupInterceptor() func() {
	// collect trace
	return trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object) (interface{}, error) {
			key := uintptr(__xgo_link_getcurg())
			localOptStack, ok := collectingMap.Load(key)
			var localOpts *collectOpts
			if ok {
				l := localOptStack.(*optStack)
				if len(l.list) > 0 {
					localOpts = l.list[len(l.list)-1]
				}
			}
			stack := &Stack{
				FuncInfo: f,
				Args:     args,
				Results:  results,
			}
			var globalRoot interface{}
			var localRoot *Root
			var initial bool
			if localOpts == nil {
				var globalLoaded bool
				globalRoot, globalLoaded = stackMap.Load(key)
				if !globalLoaded {
					initial = true
				}
			} else {
				localRoot = localOpts.root
				if localRoot == nil {
					initial = true
				}
			}
			if initial {
				// initial stack
				root := &Root{
					Top:   stack,
					Begin: time.Now(),
					Children: []*Stack{
						stack,
					},
				}
				stack.Begin = int64(time.Since(root.Begin))
				if localOpts == nil {
					stackMap.Store(key, root)
				} else {
					localOpts.root = root
				}
				// NOTE: for initial stack, the data is nil
				return nil, nil
			}
			var root *Root
			if localOpts != nil {
				root = localRoot
			} else {
				root = globalRoot.(*Root)
			}
			stack.Begin = int64(time.Since(root.Begin))
			prevTop := root.Top
			root.Top.Children = append(root.Top.Children, stack)
			root.Top = stack
			return prevTop, nil
		},
		Post: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object, data interface{}) error {
			trap.Skip()
			key := uintptr(__xgo_link_getcurg())

			localOptStack, ok := collectingMap.Load(key)
			var localOpts *collectOpts
			if ok {
				l := localOptStack.(*optStack)
				if len(l.list) > 0 {
					localOpts = l.list[len(l.list)-1]
				}
			}
			var root *Root
			if localOpts != nil {
				if localOpts.root == nil {
					panic(fmt.Errorf("unbalanced stack"))
				}
				root = localOpts.root
			} else {
				v, ok := stackMap.Load(key)
				if !ok {
					panic(fmt.Errorf("unbalanced stack"))
				}
				root = v.(*Root)
			}

			// detect panic
			pe := __xgo_link_peek_panic()
			if pe != nil {
				root.Top.Panic = true
				peErr, ok := pe.(error)
				if !ok {
					peErr = fmt.Errorf("panic: %v", pe)
				}
				root.Top.Error = peErr
			} else {
				if errObj, ok := results.(core.ObjectWithErr); ok {
					fnErr := errObj.GetErr().Value()
					if fnErr != nil {
						root.Top.Error = fnErr.(error)
					}
				}
			}
			root.Top.End = int64(time.Since(root.Begin))
			if data == nil {
				// stack finished
				if localOpts != nil {
					if localOpts.onComplete != nil {
						localOpts.onComplete(root)
						return nil
					}
					err := emitTrace(localOpts.name, root)
					if err != nil {
						return err
					}
					return nil
				}

				// global
				stackMap.Delete(key)
				err := emitTrace("", root)
				if err != nil {
					return err
				}
				return nil
			}
			// pop stack
			root.Top = data.(*Stack)
			return nil
		},
	})
}

var collectingMap sync.Map // <uintptr> -> []*collectOpts

type optStack struct {
	list []*collectOpts
}

func collect(f func(), collOpts *collectOpts) {
	if atomic.LoadInt32(&enabledGlobal) == 0 {
		cancel := setupInterceptor()
		defer cancel()
	}
	key := uintptr(__xgo_link_getcurg())
	if collOpts.name == "" {
		collOpts.name = fmt.Sprintf("g_%x", uint(key))
	}

	act, _ := collectingMap.LoadOrStore(key, &optStack{})
	opts := act.(*optStack)

	// push
	opts.list = append(opts.list, collOpts)
	defer func() {
		// pop
		opts.list = opts.list[:len(opts.list)-1]
		if len(opts.list) == 0 {
			collectingMap.Delete(key)
		}
	}()
	f()
}

func getTraceOutput() string {
	return os.Getenv("XGO_TRACE_OUTPUT")
}

var marshalStack func(root *Root) ([]byte, error)

func SetMarshalStack(fn func(root *Root) ([]byte, error)) {
	marshalStack = fn
}

func fmtStack(root *Root) (data []byte, err error) {
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
		return marshalStack(root)
	}
	return json.Marshal(root.Export())
}

// this should also be marked as trap.Skip()
// TODO: may add callback for this
func emitTrace(name string, root *Root) error {
	if name == "" {
		key := uintptr(__xgo_link_getcurg())
		tinfo, ok := testInfoMapping.Load(key)
		if ok {
			name = tinfo.(*testInfo).name
		}
	}

	xgoTraceOutput := getTraceOutput()
	useStdout := xgoTraceOutput == "stdout"
	subName := name
	if name == "" {
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
	trace, stackErr := fmtStack(root)
	if stackErr != nil {
		traceOut = []byte("error:" + stackErr.Error())
	} else {
		traceOut = trace
	}

	if useStdout {
		fmt.Printf("%s\n", traceOut)
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
