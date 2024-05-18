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

// all func infos
var funcInfos []*core.FuncInfo
var funcInfoMapping map[string]map[string]*core.FuncInfo         // pkg -> identifyName -> FuncInfo
var funcPCMapping map[uintptr]*core.FuncInfo                     // pc->FuncInfo
var varAddrMapping map[uintptr]*core.FuncInfo                    // addr->FuncInfo
var funcFullNameMapping map[string]*core.FuncInfo                // fullName -> FuncInfo
var interfaceMapping map[string]map[string]*core.FuncInfo        // pkg -> interfaceName -> FuncInfo
var typeMethodMapping map[reflect.Type]map[string]*core.FuncInfo // reflect.Type -> interfaceName -> FuncInfo

func init() {
	funcPCMapping = make(map[uintptr]*core.FuncInfo)
	funcInfoMapping = make(map[string]map[string]*core.FuncInfo)
	funcFullNameMapping = make(map[string]*core.FuncInfo)
	interfaceMapping = make(map[string]map[string]*core.FuncInfo)
	varAddrMapping = make(map[uintptr]*core.FuncInfo)

	// this will consume all staged func infos in runtime,
	// and set registerFuncInfo for later registering
	__xgo_link_retrieve_all_funcs_and_clear(registerFuncInfo)
}

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_retrieve_all_funcs_and_clear(f func(fn interface{})) {
	// linked at runtime
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_retrieve_all_funcs_and_clear(requires xgo).")
}

func __xgo_link_get_pc_name(pc uintptr) string {
	fmt.Fprintf(os.Stderr, "WARNING: failed to link __xgo_link_get_pc_name(requires xgo).\n")
	return ""
}

func GetFuncs() []*core.FuncInfo {
	return funcInfos
}

func InfoFunc(fn interface{}) *core.FuncInfo {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("given type is not a func: %T", fn))
	}
	// deref to pc
	pc := v.Pointer()
	return funcPCMapping[pc]
}

func InfoVar(addr interface{}) *core.FuncInfo {
	v := reflect.ValueOf(addr)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("given type is not a pointer: %T", addr))
	}
	ptr := v.Pointer()
	return varAddrMapping[ptr]
}

// maybe rename to FuncForPC
func InfoPC(pc uintptr) *core.FuncInfo {
	return funcPCMapping[pc]
}

// maybe rename to FuncForGeneric
func Info(pkg string, identityName string) *core.FuncInfo {
	return funcInfoMapping[pkg][identityName]
}

// GetFuncByPkg:
//
//	pkg.Func
//	pkg.Recv.Func
//	pkg.(*Recv).Func
func GetFuncByPkg(pkg string, name string) *core.FuncInfo {
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
	return getTypeMethodMapping()[typ]
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

var errType = reflect.TypeOf((*error)(nil)).Elem()
var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

func registerFuncInfo(fnInfo interface{}) {
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
	// fmt.Printf("register: %s %s\n", pkgPath, identityName)
	var fnKind core.Kind
	fnKindV := rv.FieldByName("Kind")
	if fnKindV.IsValid() {
		fnKind = core.Kind(fnKindV.Int())
	}
	varField := rv.FieldByName("Var")
	recvTypeName := rv.FieldByName("RecvTypeName").String()
	recvPtr := rv.FieldByName("RecvPtr").Bool()
	name := rv.FieldByName("Name").String()
	interface_ := rv.FieldByName("Interface").Bool()
	generic := rv.FieldByName("Generic").Bool()

	var stdlib bool
	stdlibField := rv.FieldByName("Stdlib")
	if stdlibField.IsValid() {
		stdlib = stdlibField.Bool()
	}

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
			if (closure || fnKind == core.Kind_Var || fnKind == core.Kind_VarPtr || fnKind == core.Kind_Const) && identityName != "" {
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
		Kind:         fnKind,
		FullName:     fullName,
		Pkg:          pkgPath,
		IdentityName: identityName,
		Name:         name,
		RecvType:     recvTypeName,
		RecvPtr:      recvPtr,

		Interface: interface_,
		Generic:   generic,
		Closure:   closure,
		Stdlib:    stdlib,

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
	if varField.IsValid() {
		info.Var = varField.Interface()
	}
	funcInfos = append(funcInfos, info)
	if !generic && info.PC != 0 {
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
	if fnKind == core.Kind_Var {
		if varField.IsValid() {
			varAddr := varField.Elem().Pointer()
			varAddrMapping[varAddr] = info
		}
	}
	if fullName != "" {
		funcFullNameMapping[fullName] = info
	}
}

var mappingTypeOnce sync.Once

func getTypeMethodMapping() map[reflect.Type]map[string]*core.FuncInfo {
	mappingTypeOnce.Do(initTypeMethodMapping)
	return typeMethodMapping
}

func initTypeMethodMapping() {
	typeMethodMapping = make(map[reflect.Type]map[string]*core.FuncInfo)
	for _, funcInfo := range funcInfos {
		registerTypeMethod(funcInfo)
	}
}

func registerTypeMethod(funcInfo *core.FuncInfo) {
	if funcInfo.Kind != core.Kind_Func {
		return
	}
	if funcInfo.Generic || funcInfo.Interface || funcInfo.RecvType == "" {
		return
	}
	if funcInfo.Func == nil || funcInfo.Name == "" {
		return
	}
	recvType := reflect.TypeOf(funcInfo.Func).In(0)
	methodMapping := typeMethodMapping[recvType]
	if methodMapping == nil {
		methodMapping = make(map[string]*core.FuncInfo, 1)
		typeMethodMapping[recvType] = methodMapping
	}
	methodMapping[funcInfo.Name] = funcInfo
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
