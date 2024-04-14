package trace

import (
	"context"
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

// hold goroutine stacks, keyed by goroutine ptr
var stackMap sync.Map        // uintptr(goroutine) -> *Root
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

var enabledGlobally bool
var interceptorRefCount int32

// Enable setup the trace interceptor
// if called from init, the interceptor is enabled
// globally. Otherwise locally
func Enable() func() {
	if __xgo_link_init_finished() {
		return enableLocal(nil)
	}
	enabledGlobally = true
	setupInterceptor()
	return func() {
		panic("global trace cannot be turned off")
	}
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

func Begin() func() {
	if !__xgo_link_init_finished() {
		panic("Begin cannot be called from init")
	}
	return enableLocal(&collectOpts{})
}

type collectOpts struct {
	name          string
	onComplete    func(root *Root)
	root          *Root
	exportOptions *ExportOptions
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

func (c *collectOpts) WithExport(expOpts *ExportOptions) *collectOpts {
	c.exportOptions = expOpts
	return c
}

func (c *collectOpts) Collect(f func()) {
	collect(f, c)
}

func (c *collectOpts) Begin() func() {
	return enableLocal(c)
}

func setupInterceptor() func() {
	if atomic.AddInt32(&interceptorRefCount, 1) > 1 {
		return func() {
			atomic.AddInt32(&interceptorRefCount, -1)
		}
	}
	// collect trace
	cancel := trap.AddInterceptorHead(&trap.Interceptor{
		Pre:  handleTracePre,
		Post: handleTracePost,
	})
	return func() {
		atomic.AddInt32(&interceptorRefCount, -1)
		cancel()
	}
}

func handleTracePre(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object) (interface{}, error) {
	if !__xgo_link_init_finished() {
		// do not collect trace while init
		return nil, trap.ErrSkip
	}
	key := uintptr(__xgo_link_getcurg())
	localOptStack, ok := collectingMap.Load(key)
	var localOpts *collectOpts
	if ok {
		l := localOptStack.(*optStack)
		if len(l.list) > 0 {
			localOpts = l.list[len(l.list)-1]
		}
	} else if !enabledGlobally {
		return nil, trap.ErrSkip
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
			Begin: getNow(),
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
		// this will signal Post to emit a trace
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
}

func handleTracePost(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object, data interface{}) error {
	key := uintptr(__xgo_link_getcurg())

	localOptStack, ok := collectingMap.Load(key)
	var localOpts *collectOpts
	if ok {
		l := localOptStack.(*optStack)
		if len(l.list) > 0 {
			localOpts = l.list[len(l.list)-1]
		}
	} else if !enabledGlobally {
		return nil
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
		root.Top = nil
		// stack finished
		if localOpts != nil {
			// handled by local options
			return nil
		}

		// global
		stackMap.Delete(key)
		emitTraceNoErr("", root, nil)
		return nil
	}
	// pop stack
	root.Top = data.(*Stack)
	return nil
}

var collectingMap sync.Map // <uintptr> -> []*collectOpts

type optStack struct {
	list []*collectOpts
}

func collect(f func(), collOpts *collectOpts) {
	finish := enableLocal(collOpts)
	defer finish()
	f()
}

func enableLocal(collOpts *collectOpts) func() {
	if collOpts == nil {
		collOpts = &collectOpts{}
	}
	cancel := setupInterceptor()
	key := uintptr(__xgo_link_getcurg())
	if collOpts.name == "" {
		var name string
		tinfo, ok := testInfoMapping.Load(key)
		if ok {
			name = tinfo.(*testInfo).name
		}
		if name == "" {
			name = fmt.Sprintf("g_%x", uint(key))
		}
		collOpts.name = name
	}
	if collOpts.root == nil {
		collOpts.root = &Root{
			Top:   &Stack{},
			Begin: getNow(),
		}
	}
	top := collOpts.root.Top

	act, _ := collectingMap.LoadOrStore(key, &optStack{})
	opts := act.(*optStack)

	// push
	opts.list = append(opts.list, collOpts)
	return func() {
		if key != uintptr(__xgo_link_getcurg()) {
			panic("finish trace from another goroutine!")
		}
		cancel()
		// pop
		opts.list = opts.list[:len(opts.list)-1]
		if len(opts.list) == 0 {
			collectingMap.Delete(key)
		}

		root := collOpts.root
		root.Children = top.Children
		root.Top = nil
		// root.Children =
		// call complete
		if collOpts.onComplete != nil {
			collOpts.onComplete(root)
		} else {
			emitTraceNoErr(collOpts.name, root, collOpts.exportOptions)
		}
	}
}

var traceOutput = os.Getenv("XGO_TRACE_OUTPUT")

func getTraceOutput() string {
	return traceOutput
}

var marshalStack func(root *Root) ([]byte, error)

func SetMarshalStack(fn func(root *Root) ([]byte, error)) {
	marshalStack = fn
}

func fmtStack(root *Root, opts *ExportOptions) (data []byte, err error) {
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
	exportRoot := root.Export(opts)
	if opts != nil {
		if opts.FilterRoot != nil {
			exportRoot = opts.FilterRoot(exportRoot)
		}
		if opts.MarshalRoot != nil {
			return opts.MarshalRoot(exportRoot)
		}
	}
	return MarshalAnyJSON(exportRoot)
}

func emitTraceNoErr(name string, root *Root, opts *ExportOptions) {
	emitTrace(name, root, opts)
}

func getNow() (now time.Time) {
	trap.Direct(func() {
		now = time.Now()
	})
	return
}
func formatTime(t time.Time, layout string) (output string) {
	trap.Direct(func() {
		output = t.Format(layout)
	})
	return
}

// this should also be marked as trap.Skip()
// TODO: may add callback for this
func emitTrace(name string, root *Root, opts *ExportOptions) error {
	xgoTraceOutput := getTraceOutput()
	if xgoTraceOutput == "off" {
		return nil
	}
	useStdout := xgoTraceOutput == "stdout"
	subName := name
	if name == "" {
		traceIDNum := int64(1)
		ghex := fmt.Sprintf("g_%x", __xgo_link_getcurg())
		traceID := "t_" + strconv.FormatInt(traceIDNum, 10)
		if xgoTraceOutput == "" {
			traceDir := formatTime(getNow(), "trace_20060102_150405")
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
	trace, stackErr := fmtStack(root, opts)
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
	trap.Direct(func() {
		err = WriteFile(subFile, traceOut, 0755)
	})
	return err
}
