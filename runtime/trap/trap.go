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
		// do not capture trap before init finished
		if __xgo_link_init_finished() {
			__xgo_link_set_trap(trapFunc)
			__xgo_link_set_trap_var(trapVar)
		} else {
			// deferred
			__xgo_link_on_init_finished(func() {
				__xgo_link_set_trap(trapFunc)
				__xgo_link_set_trap_var(trapVar)
			})
		}
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

var trappingMark sync.Map // <goroutine key> -> struct{}{}
var trappingPC sync.Map   // <gorotuine key> -> PC

var inspectingMap sync.Map // <goroutine key> -> interceptor

// link to runtime
// xgo:notrap
func trapFunc(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	if isByPassing() {
		return nil, false
	}
	inspectingFn, inspecting := inspectingMap.Load(uintptr(__xgo_link_getcurg()))
	var interceptors []*Interceptor
	var group int
	var n int
	if !inspecting {
		// never trap any function from runtime
		interceptors, group = getAllInterceptors()
		n = len(interceptors)
		if n == 0 {
			return nil, false
		}
	}

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

	// setup context
	setTrappingPC(pc)
	defer clearTrappingPC()
	return trap(f, interceptors, group, recv, args, results)
}

func trapVar(pkgPath string, name string, tmpVarAddr interface{}, takeAddr bool) {
	if isByPassing() {
		return
	}
	interceptors, group := getAllInterceptors()
	n := len(interceptors)
	if n == 0 {
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
	post, _ := trap(fnInfo, interceptors, group, nil, nil, []interface{}{tmpVarAddr})
	if post != nil {
		// NOTE: must in defer, because in post we
		// may capture panic
		defer post()
	}
}
func trap(f *core.FuncInfo, interceptors []*Interceptor, group int, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	dispose := setTrappingMark(group)
	if dispose == nil {
		return nil, false
	}
	defer dispose()

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
	n := len(interceptors)
	dataList := make([]interface{}, n)
	for i := n - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		if interceptor.Pre == nil {
			continue
		}
		// if
		data, err := interceptor.Pre(ctx, f, req, resObject)
		dataList[i] = data
		if err != nil {
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
	// always run post in defer
	return func() {
		dispose := setTrappingMark(group)
		if dispose == nil {
			return
		}
		defer dispose()

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
	val, ok := trappingPC.Load(key)
	if !ok {
		return 0
	}
	return val.(uintptr)
}

type trappingGroup struct {
	m map[int]bool
}

func setTrappingMark(group int) func() {
	key := uintptr(__xgo_link_getcurg())
	g := &trappingGroup{}
	v, loaded := trappingMark.LoadOrStore(key, g)
	if loaded {
		g = v.(*trappingGroup)
		if g.m[group] {
			return nil
		}
	} else {
		g.m = make(map[int]bool, 1)
	}
	g.m[group] = true
	return func() {
		g.m[group] = false
	}
}

func clearTrappingMarkAllGroup() {
	key := uintptr(__xgo_link_getcurg())
	trappingMark.Delete(key)
}

func setTrappingPC(pc uintptr) {
	key := uintptr(__xgo_link_getcurg())
	trappingPC.Store(key, pc)
}

func clearTrappingPC() {
	key := uintptr(__xgo_link_getcurg())
	trappingPC.Delete(key)
}
