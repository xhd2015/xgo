package functab

import (
	"context"
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
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_retrieve_all_funcs_and_clear.(xgo required)")
}

func __xgo_link_on_init_finished(f func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_init_finished.(xgo required)")
}

func __xgo_link_get_pc_name(pc uintptr) string {
	fmt.Fprintf(os.Stderr, "WARNING: failed to link __xgo_link_get_pc_name(requires xgo).\n")
	return ""
}

var funcInfos []*core.FuncInfo
var funcInfoMapping map[string]map[string]*core.FuncInfo         // pkg -> identifyName -> FuncInfo
var funcPCMapping map[uintptr]*core.FuncInfo                     // pc->FuncInfo
var funcFullNameMapping map[string]*core.FuncInfo                // fullName -> FuncInfo
var interfaceMapping map[string]map[string]*core.FuncInfo        // pkg -> interfaceName -> FuncInfo
var typeMethodMapping map[reflect.Type]map[string]*core.FuncInfo // reflect.Type -> interfaceName -> FuncInfo

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

// GetFuncByPkg:
//
//	pkg.Func
//	pkg.Recv.Func
//	pkg.(*Recv).Func
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
	// already have "("
	if strings.HasPrefix(name, "(") {
		return nil
	}

	// NOTE: may handle generic
	dotIdx := strings.Index(name, ".")
	if dotIdx < 0 {
		return nil
	}
	typName := name[:dotIdx]
	funcName := name[dotIdx+1:]

	return pkgMapping["(*"+typName+")."+funcName]
}

func GetFuncByFullName(fullName string) *core.FuncInfo {
	f := funcFullNameMapping[fullName]
	if f != nil {
		return f
	}
	return getInterfaceOrGenericByFullName(fullName)
}

func GetTypeMethods(typ reflect.Type) map[string]*core.FuncInfo {
	ensureTypeMapping()
	return typeMethodMapping[typ]
}

func getInterfaceOrGenericByFullName(fullName string) *core.FuncInfo {
	pkgPath, recvName, recvPtr, typeGeneric, funcGeneric, funcName := core.ParseFuncName(fullName)
	if typeGeneric != "" || funcGeneric != "" {
		// generic, currently does not solve generic interface
		idName := funcName
		if recvName != "" {
			idName = recvName + "." + funcName
		}
		return GetFuncByPkg(pkgPath, idName)
	}
	if recvName != "" && !recvPtr {
		// if the recv is a pointer, it cannot be interface
		intfMethod := interfaceMapping[pkgPath][recvName]
		if intfMethod != nil {
			return intfMethod
		}
	}
	return nil
}

var mappingOnce sync.Once

var errType = reflect.TypeOf((*error)(nil)).Elem()
var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

func ensureMapping() {
	mappingOnce.Do(func() {
		funcPCMapping = make(map[uintptr]*core.FuncInfo)
		funcInfoMapping = make(map[string]map[string]*core.FuncInfo)
		funcFullNameMapping = make(map[string]*core.FuncInfo)
		interfaceMapping = make(map[string]map[string]*core.FuncInfo)
		__xgo_link_retrieve_all_funcs_and_clear(func(fnInfo interface{}) {
			rv := reflect.ValueOf(fnInfo)
			if rv.Kind() != reflect.Struct {
				panic(fmt.Errorf("expect struct, actual: %s", rv.Kind().String()))
			}
			closure := rv.FieldByName("Closure").Bool()
			identityName := rv.FieldByName("IdentityName").String()
			if identityName == "" {
				if !closure {
					return
				}
				// 	fmt.Fprintf(os.Stderr, "empty name\n",pkgPath)
			}

			pkgPath := rv.FieldByName("PkgPath").String()
			recvTypeName := rv.FieldByName("RecvTypeName").String()
			recvPtr := rv.FieldByName("RecvPtr").Bool()
			name := rv.FieldByName("Name").String()
			interface_ := rv.FieldByName("Interface").Bool()
			generic := rv.FieldByName("Generic").Bool()
			f := rv.FieldByName("Fn").Interface()

			var firstArgCtx bool
			var lastResErr bool
			var pc uintptr
			var fullName string
			if !generic && !interface_ {
				if f != nil {
					// TODO: move all ctx, err check logic here
					ft := reflect.TypeOf(f)
					off := 0
					if recvTypeName != "" {
						off = 1
					}
					if ft.NumIn() > off && ft.In(off).Implements(ctxType) {
						firstArgCtx = true
					}
					// NOTE: use == instead of implements
					if ft.NumOut() > 0 && ft.Out(ft.NumOut()-1) == errType {
						lastResErr = true
					}
					pc = getFuncPC(f)
					fullName = __xgo_link_get_pc_name(pc)
				} else {
					if closure && identityName != "" {
						fullName = pkgPath + "." + identityName
					}
				}
			}
			recvName := rv.FieldByName("RecvName").String()
			argNames := rv.FieldByName("ArgNames").Interface().([]string)
			resNames := rv.FieldByName("ResNames").Interface().([]string)
			file := rv.FieldByName("File").String()
			line := int(rv.FieldByName("Line").Int())

			// debug
			// fmt.Printf("reg: %s\n", fullName)
			// if pkgPath == "main" {
			// 	fmt.Fprintf(os.Stderr, "reg: funcName=%s,pc=%x,generic=%v,genericname=%s\n", funcName, pc, generic, genericName)
			// }
			// _, recvTypeName, recvPtr, name := core.ParseFuncName(identityName, false)
			info := &core.FuncInfo{
				FullName:     fullName,
				Pkg:          pkgPath,
				IdentityName: identityName,
				Name:         name,
				RecvType:     recvTypeName,
				RecvPtr:      recvPtr,

				Interface: interface_,
				Generic:   generic,
				Closure:   closure,

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
			if interface_ && recvTypeName != "" {
				pkgMapping := interfaceMapping[pkgPath]
				if pkgMapping == nil {
					pkgMapping = make(map[string]*core.FuncInfo, 1)
					interfaceMapping[pkgPath] = pkgMapping
				}
				pkgMapping[recvTypeName] = info
			}
			if fullName != "" {
				funcFullNameMapping[fullName] = info
			}
		})
	})
}

var mappingTypeOnce sync.Once

func ensureTypeMapping() {
	ensureMapping()
	mappingTypeOnce.Do(func() {
		typeMethodMapping = make(map[reflect.Type]map[string]*core.FuncInfo)
		for _, funcInfo := range funcInfos {
			if funcInfo.Generic || funcInfo.Interface || funcInfo.RecvType == "" {
				continue
			}
			if funcInfo.Func == nil || funcInfo.Name == "" {
				continue
			}
			recvType := reflect.TypeOf(funcInfo.Func).In(0)
			methodMapping := typeMethodMapping[recvType]
			if methodMapping == nil {
				methodMapping = make(map[string]*core.FuncInfo, 1)
				typeMethodMapping[recvType] = methodMapping
			}
			methodMapping[funcInfo.Name] = funcInfo
		}
	})
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
