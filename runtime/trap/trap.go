package trap

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

var setupOnce sync.Once

func ensureTrapInstall() {
	setupOnce.Do(func() {
		// set trap once needed, no matter it
		// is inside init or not
		__xgo_link_set_trap(trapFunc)
		__xgo_link_set_trap_var(trapVar)

		// // do not capture trap before init finished
		// if __xgo_link_init_finished() {
		// } else {
		// 	// deferred
		// 	__xgo_link_on_init_finished(func() {
		// 		__xgo_link_set_trap(trapFunc)
		// 		__xgo_link_set_trap_var(trapVar)
		// 	})
		// }
	})
}

func init() {
	__xgo_link_on_gonewproc(func(g uintptr) {
		if isByPassing() {
			return
		}
		local := getLocalInterceptorList()
		if local == nil || (len(local.head) == 0 && len(local.tail) == 0) {
			return
		}
		// inherit interceptors of last group
		localInterceptors.Store(g, &interceptorGroup{
			groups: []*interceptorList{{
				list: local.copy(),
			}},
		})
	})
}

func __xgo_link_set_trap(trapImpl func(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_set_trap(requires xgo).")
}

func __xgo_link_set_trap_var(trap func(pkgPath string, name string, tmpVarAddr interface{}, takeAddr bool)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_set_trap_var(requires xgo).")
}
func __xgo_link_on_gonewproc(f func(g uintptr)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_gonewproc(requires xgo).")
}
func __xgo_link_on_init_finished(f func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_init_finished(requires xgo).")
}

// Skip serves as mark to tell xgo not insert
// trap instructions for the function that
// calls Skip()
// NOTE: the function body is intenionally leave empty
// as trap.Skip() is just a mark that makes
// sense at compile time.
func Skip() {}

var stackMapping sync.Map // <goroutine key> -> root

var inspectingMap sync.Map // <goroutine key> -> interceptor

type root struct {
	top          *stack
	intercepting bool // is executing intercepting? to avoid re-entrance
}

type stack struct {
	parent   *stack
	funcInfo *core.FuncInfo
	stage    stage
	pc       uintptr // the actual pc
}

type stage int

const (
	stage_pre     stage = 0
	stage_execute stage = 1
	stage_post    stage = 2
)

// link to runtime
// xgo:notrap
func trapFunc(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	if isByPassing() {
		return nil, false
	}
	inspectingFn, inspecting := inspectingMap.Load(uintptr(__xgo_link_getcurg()))

	// NOTE: this may return nil for generic template
	var f *core.FuncInfo
	if !generic {
		f = functab.InfoPC(pc)
	}
	if f == nil {
		// generic and closure under go1.22
		f = functab.Info(pkgPath, identityName)
	}
	if f == nil {
		// no func found
		return nil, false
	}
	if f.RecvType != "" && methodHasBeenTrapped && recv == nil {
		// method
		// let go to the next interceptor
		return nil, false
	}
	if inspecting {
		inspectingFn.(inspectingFunc)(f, recv, pc)
		// abort,never really call the target
		return nil, true
	}
	// f is an interceptor or ignored by user
	if funcIgnored(f) {
		return nil, false
	}

	return trap(f, pc, recv, args, results)
}

func trapVar(pkgPath string, name string, tmpVarAddr interface{}, takeAddr bool) {
	if isByPassing() {
		return
	}
	identityName := name
	if takeAddr {
		identityName = "*" + name
	}
	fnInfo := functab.Info(pkgPath, identityName)
	if fnInfo == nil {
		return
	}
	if fnInfo.Kind != core.Kind_Var && fnInfo.Kind != core.Kind_VarPtr && fnInfo.Kind != core.Kind_Const {
		return
	}
	// NOTE: stop always ignored because this is a simple get
	post, _ := trap(fnInfo, 0, nil, nil, []interface{}{tmpVarAddr})
	if post != nil {
		// NOTE: must in defer, because in post we
		// may capture panic
		defer post()
	}
}
func trap(f *core.FuncInfo, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	// never trap any function from runtime
	key := uintptr(__xgo_link_getcurg())
	r := &root{}
	rv, loaded := stackMapping.LoadOrStore(key, r)
	if loaded {
		r = rv.(*root)
	}
	// fmt.Printf("trap: %s.%s intercepting=%v\n", f.Pkg, f.IdentityName, r.intercepting)
	interceptors, _ := getAllInterceptors(f, !r.intercepting)
	n := len(interceptors)
	if n == 0 {
		return nil, false
	}

	var resetFlag bool
	parent := r.top
	if !r.intercepting {
		resetFlag = true
		r.intercepting = true
		defer func() {
			r.intercepting = false
		}()
	}
	stack := &stack{
		parent:   parent,
		funcInfo: f,
		stage:    stage_pre,
		pc:       pc,
	}
	r.top = stack

	// dispose := setTrappingMark(group, f)
	// if dispose == nil {
	// 	return nil, false
	// }
	// defer dispose()

	// retrieve context
	var ctx context.Context
	if f.FirstArgCtx {
		// TODO: is *HttpRequest a *Context?

		// NOTE: ctx can be nil when doing InspectPC
		argCtx := reflect.ValueOf(args[0]).Elem().Interface()
		if argCtx != nil {
			ctx = argCtx.(context.Context)
		}
		// ctx = *(args[0].(*context.Context))
	} else if f.Closure {
		if len(args) > 0 {
			argCtx, ok := reflect.ValueOf(args[0]).Elem().Interface().(context.Context)
			if ok {
				// modify on the fly
				f.FirstArgCtx = true
				ctx = argCtx
			}
		}
	}
	var perr *error
	if f.LastResultErr {
		perr = results[len(results)-1].(*error)
	} else if f.Closure {
		if len(results) > 0 {
			resErr, ok := reflect.ValueOf(results[0]).Interface().(*error)
			if ok {
				f.LastResultErr = true
				perr = resErr
			}
		}
	}

	// TODO: set FirstArgCtx and LastResultErr
	req := make(object, 0, len(args))
	result := make(object, 0, len(results))
	if f.RecvType != "" {
		req = append(req, field{
			name:   f.RecvName,
			valPtr: recv,
		})
	}
	if !f.FirstArgCtx {
		req = appendFields(req, args, f.ArgNames)
	} else {
		argNames := f.ArgNames
		if argNames != nil {
			argNames = argNames[1:]
		}
		req = appendFields(req, args[1:], argNames)
	}

	var resObject core.Object
	if !f.LastResultErr {
		result = appendFields(result, results, f.ResNames)
		resObject = result
	} else {
		resNames := f.ResNames
		var errName string
		if resNames != nil {
			errName = resNames[len(resNames)-1]
			resNames = resNames[:len(resNames)-1]
		}
		errField := field{
			name:   errName,
			valPtr: results[len(results)-1],
		}
		result = appendFields(result, results[:len(results)-1], resNames)
		resObject = &objectWithErr{
			object: result,
			err:    errField,
		}
	}

	// TODO: what about inlined func?
	// funcArgs := &FuncArgs{
	// 	Recv:    recv,
	// 	Args:    args,
	// 	Results: results,
	// }

	// NOTE: context.TODO() is a constant
	if ctx == nil {
		ctx = context.TODO()
	}

	var firstPreErr error

	abortIdx := -1
	dataList := make([]interface{}, n)
	skipIndex := make([]bool, n)
	for i := n - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		if interceptor.Pre == nil {
			continue
		}
		// if
		data, err := interceptor.Pre(ctx, f, req, resObject)
		dataList[i] = data
		if err != nil {
			if err == ErrSkip {
				skipIndex[i] = true
				continue
			}
			// always break on error
			firstPreErr = err
			abortIdx = i
			break
		}
	}

	// not really an error
	if firstPreErr == ErrAbort {
		firstPreErr = nil
	}
	stack.stage = stage_execute
	// always run post in defer
	return func() {
		stack.stage = stage_post
		defer func() {
			r.top = parent
		}()

		if resetFlag {
			r.intercepting = true
			defer func() {
				r.intercepting = false
			}()
		}

		var lastPostErr error = firstPreErr
		idx := 0
		if abortIdx != -1 {
			idx = abortIdx
		}
		for i := idx; i < n; i++ {
			interceptor := interceptors[i]
			if interceptor.Post == nil {
				continue
			}
			if skipIndex[i] {
				continue
			}
			err := interceptor.Post(ctx, f, req, resObject, dataList[i])
			if err != nil {
				if err == ErrAbort {
					return
				}
				lastPostErr = err
			}
		}
		if lastPostErr != nil {
			if perr != nil {
				*perr = lastPostErr
			} else {
				panic(lastPostErr)
			}
		}
	}, abortIdx != -1
}

func GetTrappingPC() uintptr {
	key := uintptr(__xgo_link_getcurg())
	val, ok := stackMapping.Load(key)
	if !ok {
		return 0
	}
	top := val.(*root).top
	if top == nil {
		return 0
	}
	return top.pc
}

func clearLocalInterceptorsAndMark() {
	key := uintptr(__xgo_link_getcurg())
	localInterceptors.Delete(key)
	bypassMapping.Delete(key)

	stackMapping.Delete(key)
}
