package functab

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
)

const __XGO_SKIP_TRAP = true

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_for_each_func(f func(pkgPath string, funcName string, pc uintptr, fn interface{}, generic bool, genericName string, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool)) {
	panic("failed to link __xgo_link_for_each_func")
}

var funcInfos []*core.FuncInfo
var funcInfoMapping map[string]*core.FuncInfo
var funcPCMapping map[uintptr]*core.FuncInfo         // pc->FuncInfo
var funcInfoGenericMapping map[string]*core.FuncInfo // generic name -> FuncInfo

func GetFuncs() []*core.FuncInfo {
	ensureMapping()
	return funcInfos
}

func GetFunc(fullName string) *core.FuncInfo {
	ensureMapping()
	return funcInfoMapping[fullName]
}

func Info(fn interface{}) *core.FuncInfo {
	ensureMapping()
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("given type is not a func: %T", fn))
	}
	// deref to pc
	pc := v.Pointer()
	return funcPCMapping[pc]
}

func InfoPC(pc uintptr) *core.FuncInfo {
	ensureMapping()
	return funcPCMapping[pc]
}

func InfoGeneric(genericName string) *core.FuncInfo {
	ensureMapping()
	return funcInfoGenericMapping[genericName]
}

func GetFuncByPkg(pkgPath string, name string) *core.FuncInfo {
	ensureMapping()
	fn := funcInfoMapping[pkgPath+"."+name]
	if fn != nil {
		return fn
	}
	dotIdx := strings.Index(name, ".")
	if dotIdx < 0 {
		return fn
	}
	typName := name[:dotIdx]
	funcName := name[dotIdx+1:]

	return funcInfoMapping[pkgPath+".(*"+typName+")."+funcName]
}

var mappingOnce sync.Once

func ensureMapping() {
	mappingOnce.Do(func() {
		funcInfoMapping = map[string]*core.FuncInfo{}
		funcPCMapping = make(map[uintptr]*core.FuncInfo)
		funcInfoGenericMapping = make(map[string]*core.FuncInfo)
		__xgo_link_for_each_func(func(pkgPath string, funcName string, pc uintptr, fn interface{}, generic bool, genericName string, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool) {
			// if pkgPath == "main" {
			// 	fmt.Fprintf(os.Stderr, "reg: funcName=%s,pc=%x,generic=%v,genericname=%s\n", funcName, pc, generic, genericName)
			// }
			var parseName string
			if !generic {
				// the funcName is in the readonly area
				// we copy it
				bytes := make([]byte, len(funcName))
				for i, b := range []byte(funcName) {
					bytes[i] = b
				}
				funcName = string(bytes)
				// prefix := pkgPath + "."
				parseName = funcName[len(pkgPath)+1:]
			} else {
				// if genericName == "" {
				// 	fmt.Fprintf(os.Stderr, "found empty generic: funcName=%s,generic=%v,genericname=%s\n", funcName, generic, genericName)
				// }
				parseName = genericName
			}
			_, recvTypeName, recvPtr, name := core.ParseFuncName(parseName, false)
			info := &core.FuncInfo{
				FullName: funcName,
				Name:     name,
				Pkg:      pkgPath,
				RecvType: recvTypeName,
				RecvPtr:  recvPtr,
				Generic:  generic,

				// runtime info
				PC:       pc, // nil for generic
				Func:     fn, // nil for geneirc
				RecvName: recvName,
				ArgNames: argNames,
				ResNames: resNames,
			}
			funcInfos = append(funcInfos, info)
			if !generic {
				funcInfoMapping[info.FullName] = info
				funcPCMapping[info.PC] = info
			} else if genericName != "" {
				funcInfoGenericMapping[genericName] = info
			}
		})
	})
}
