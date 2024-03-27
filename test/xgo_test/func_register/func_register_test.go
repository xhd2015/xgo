package func_register

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func init() {
	__xgo_link_on_init_finished(initFuncMapping)
}

func __xgo_link_retrieve_all_funcs_and_clear(f func(fn interface{})) {
	panic("should be linked by compiler")
}

func __xgo_link_on_init_finished(f func()) {
	panic("should be linked by compiler")
}

func __xgo_link_get_pc_name(pc uintptr) string {
	panic("show be linked by compiler")
}

var errType = reflect.TypeOf((*error)(nil)).Elem()
var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

type fnInfo struct {
	FullName     string
	PC           uintptr
	PkgPath      string
	IdentityName string

	Closure bool

	ArgNames []string
	ResNames []string

	FirstArgCtx bool
	LastResErr  bool
}

var fnInfos []*fnInfo

func initFuncMapping() {
	__xgo_link_retrieve_all_funcs_and_clear(func(fn interface{}) {
		rv := reflect.ValueOf(fn)
		pkgPath := rv.FieldByName("PkgPath").String()
		idName := rv.FieldByName("IdentityName").String()

		interface_ := rv.FieldByName("Interface").Bool()
		generic := rv.FieldByName("Generic").Bool()
		closure_ := rv.FieldByName("Closure").Bool()
		argNames, _ := rv.FieldByName("ArgNames").Interface().([]string)
		resNames, _ := rv.FieldByName("ResNames").Interface().([]string)
		f := rv.FieldByName("Fn").Interface()
		var pc uintptr
		var fullName string

		var firstArgCtx bool
		var lastResErr bool
		if !generic && !interface_ {
			ft := reflect.TypeOf(f)
			if ft.NumIn() > 0 && ft.In(0).Implements(ctxType) {
				firstArgCtx = true
			}
			if ft.NumOut() > 0 && ft.Out(ft.NumOut()-1).Implements(errType) {
				lastResErr = true
			}
			pc = getFuncPC(f)
			fullName = __xgo_link_get_pc_name(pc)
		}
		// debug
		if false {
			fmt.Printf("fn: %s\n", fullName)
			fmt.Printf("   closure_: %v\n", closure_)
			fmt.Printf("   argNames: %v\n", argNames)
			fmt.Printf("   resNames: %v\n", resNames)
			fmt.Printf("   firstArgCtx: %v\n", firstArgCtx)
			fmt.Printf("   lastResErr: %v\n", lastResErr)
			fmt.Printf("\n")
		}

		fnInfos = append(fnInfos, &fnInfo{
			FullName:     fullName,
			PC:           pc,
			Closure:      closure_,
			PkgPath:      pkgPath,
			ArgNames:     argNames,
			ResNames:     resNames,
			IdentityName: idName,
			FirstArgCtx:  firstArgCtx,
			LastResErr:   lastResErr,
		})
	})
}

// TODO: add test on closure register
// go run ./cmd/xgo test -v -run TestFuncRegister ./test/xgo_test/func_register
// go run ./script/run-test/ --include go1.18.10 --xgo-test-only -run TestFuncRegister -v ./test/xgo_test/func_register
func TestFuncRegister(t *testing.T) {
	var hasF bool
	var hasClosure1 bool

	var closure1Fn *fnInfo

	closure1PC := reflect.ValueOf(closure1_).Pointer()
	closureName := __xgo_link_get_pc_name(closure1PC)

	e1 := reflect.TypeOf((*Err)(nil)).Implements(errType)
	if !e1 {
		t.Fatalf("expect *Err implements error")
	}
	e2 := reflect.TypeOf((*Err)(nil)).Elem().Implements(errType)
	if !e2 {
		t.Fatalf("expect Err implements error")
	}

	c1 := reflect.TypeOf((*Ctx)(nil)).Implements(ctxType)
	if !c1 {
		t.Fatalf("expect *Ctx implements context.Context")
	}
	c2 := reflect.TypeOf((*Ctx)(nil)).Elem().Implements(ctxType)
	if !c2 {
		t.Fatalf("expect Ctx implements context.Context")
	}

	// debug
	// t.Logf("closureName: %s", closureName)
	for _, fn := range fnInfos {
		// debug
		// t.Logf("fn: %v %v", fn.PkgPath, fn.IdentityName)
		if fn.PkgPath == "github.com/xhd2015/xgo/test/xgo_test/func_register" && fn.IdentityName == "F" {
			hasF = true
		}
		if fn.FullName == closureName {
			hasClosure1 = true
			closure1Fn = fn
		}
	}
	if !hasF {
		t.Fatalf("expect has F, actual not found")
	}
	if !hasClosure1 {
		t.Fatalf("expect has clousre1, actual not found")
	}
	if !closure1Fn.Closure {
		t.Fatalf("expect clousre1 set Closure")
	}
	if !closure1Fn.FirstArgCtx {
		t.Fatalf("expect clousre1 first arg ctx")
	}
	if !closure1Fn.LastResErr {
		t.Fatalf("expect clousre1 last res error")
	}
}

func F() {}

var closure1_ = func(ctx Ctx, a string) *Err {
	return nil
}

func _() {
	var c context.Context = Ctx{}
	var c2 context.Context = &Ctx{}

	var e1 error = Err{}
	var e2 error = &Err{}

	_ = c
	_ = c2
	_ = e1
	_ = e2
}
func getFuncPC(fn interface{}) uintptr {
	type _func struct {
		pc uintptr
	}
	type _intf struct {
		_    uintptr
		data *_func
	}
	return ((*_intf)(unsafe.Pointer(&fn))).data.pc
	// v := (*_intf)(unsafe.Pointer(&fn))
	// return v.data.pc
}

type Ctx struct {
	context.Context
}
type Err struct {
	error
}
