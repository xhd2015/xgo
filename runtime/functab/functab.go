package functab

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
)

const __XGO_SKIP_TRAP = true

func init() {
	func() {
		defer func() {
			if e := recover(); e != nil {
				if s, ok := e.(string); ok && s == "failed to link __xgo_link_on_init_finished" {
					// silent as this is not always needed to run eagerly
					return
				}
				panic(e)
			}
		}()
		__xgo_link_on_init_finished(ensureMapping)
	}()
}

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_for_each_func(f func(pkgPath string, recvTypeName string, recvPtr bool, name string, identityName string, generic bool, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool)) {
	panic("failed to link __xgo_link_for_each_func")
}

func __xgo_link_on_init_finished(f func()) {
	panic("failed to link __xgo_link_on_init_finished")
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
		__xgo_link_for_each_func(func(pkgPath string, recvTypeName string, recvPtr bool, name string, identityName string, generic bool, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool) {
			if identityName == "" {
				// 	fmt.Fprintf(os.Stderr, "empty name\n",pkgPath)
				return
			}
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

				// runtime info
				PC:       pc, // nil for generic
				Func:     fn, // nil for geneirc
				RecvName: recvName,
				ArgNames: argNames,
				ResNames: resNames,
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
