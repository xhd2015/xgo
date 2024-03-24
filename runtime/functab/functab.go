package functab

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
)

const __XGO_SKIP_TRAP = true

func init() {
	__xgo_link_on_init_finished(ensureMapping)
}

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_retrieve_all_funcs_and_clear(f func(fn interface{})) {
	// linked at runtime
	fmt.Fprintln(os.Stderr, "failed to link __xgo_link_retrieve_all_funcs_and_clear")
}

func __xgo_link_on_init_finished(f func()) {
	fmt.Fprintln(os.Stderr, "failed to link __xgo_link_on_init_finished")
}

var funcInfos []*core.FuncInfo
var funcInfoMapping map[string]map[string]*core.FuncInfo // pkg -> identifyName -> FuncInfo
var funcPCMapping map[uintptr]*core.FuncInfo             // pc->FuncInfo

func GetFuncs() []*core.FuncInfo {
	ensureMapping()
	return funcInfos
}

func InfoFunc(fn interface{}) *core.FuncInfo {
	ensureMapping()
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("given type is not a func: %T", fn))
	}
	// deref to pc
	pc := v.Pointer()
	return funcPCMapping[pc]
}

// maybe rename to FuncForPC
func InfoPC(pc uintptr) *core.FuncInfo {
	ensureMapping()
	return funcPCMapping[pc]
}

// maybe rename to FuncForGeneric
func Info(pkg string, identityName string) *core.FuncInfo {
	ensureMapping()
	return funcInfoMapping[pkg][identityName]
}

func GetFuncByPkg(pkg string, name string) *core.FuncInfo {
	ensureMapping()
	pkgMapping := funcInfoMapping[pkg]
	if pkgMapping == nil {
		return nil
	}
	fn := pkgMapping[name]
	if fn != nil {
		return fn
	}
	dotIdx := strings.Index(name, ".")
	if dotIdx < 0 {
		return fn
	}
	typName := name[:dotIdx]
	funcName := name[dotIdx+1:]

	return pkgMapping[".(*"+typName+")."+funcName]
}

var mappingOnce sync.Once

func ensureMapping() {
	mappingOnce.Do(func() {
		funcPCMapping = make(map[uintptr]*core.FuncInfo)
		funcInfoMapping = make(map[string]map[string]*core.FuncInfo)
		__xgo_link_retrieve_all_funcs_and_clear(func(fnInfo interface{}) {
			rv := reflect.ValueOf(fnInfo)
			if rv.Kind() != reflect.Struct {
				panic(fmt.Errorf("expect struct, actual: %s", rv.Kind().String()))
			}
			identityName := rv.FieldByName("IdentityName").String()
			if identityName == "" {
				// 	fmt.Fprintf(os.Stderr, "empty name\n",pkgPath)
				return
			}

			pkgPath := rv.FieldByName("PkgPath").String()
			recvTypeName := rv.FieldByName("RecvTypeName").String()
			recvPtr := rv.FieldByName("RecvPtr").Bool()
			name := rv.FieldByName("Name").String()
			generic := rv.FieldByName("Generic").Bool()
			f := rv.FieldByName("Fn").Interface()
			var pc uintptr
			if !generic {
				pc = getFuncPC(f)
			}
			recvName := rv.FieldByName("RecvName").String()
			argNames := rv.FieldByName("ArgNames").Interface().([]string)
			resNames := rv.FieldByName("ResNames").Interface().([]string)
			firstArgCtx := rv.FieldByName("FirstArgCtx").Bool()
			lastResErr := rv.FieldByName("LastResErr").Bool()
			file := rv.FieldByName("File").String()
			line := int(rv.FieldByName("Line").Int())
			// if pkgPath == "main" {
			// 	fmt.Fprintf(os.Stderr, "reg: funcName=%s,pc=%x,generic=%v,genericname=%s\n", funcName, pc, generic, genericName)
			// }
			// _, recvTypeName, recvPtr, name := core.ParseFuncName(identityName, false)
			info := &core.FuncInfo{
				Pkg:          pkgPath,
				IdentityName: identityName,
				Name:         name,
				RecvType:     recvTypeName,
				RecvPtr:      recvPtr,
				Generic:      generic,

				File: file,
				Line: line,

				// runtime info
				PC:   pc, // nil for generic
				Func: f,  // nil for geneirc

				RecvName: recvName,
				ArgNames: argNames,
				ResNames: resNames,

				// brief info
				FirstArgCtx:   firstArgCtx,
				LastResultErr: lastResErr,
			}
			funcInfos = append(funcInfos, info)
			if !generic {
				funcPCMapping[info.PC] = info
			}
			if identityName != "" {
				pkgMapping := funcInfoMapping[pkgPath]
				if pkgMapping == nil {
					pkgMapping = make(map[string]*core.FuncInfo, 1)
					funcInfoMapping[pkgPath] = pkgMapping
				}
				pkgMapping[identityName] = info
			}
		})
	})
}

func getFuncPC(fn interface{}) uintptr {
	type intf struct {
		_  uintptr
		pc *uintptr
	}
	v := (*intf)(unsafe.Pointer(&fn))
	return *v.pc
}
