package functab

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/core/info"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// all func infos
var funcInfos []*info.Func
var funcInfoMapping map[string]map[string]*info.Func         // pkg -> identifyName -> FuncInfo
var funcPCMapping map[uintptr]*info.Func                     // pc->FuncInfo
var varAddrMapping map[uintptr]*info.Func                    // addr->FuncInfo
var funcFullNameMapping map[string]*info.Func                // fullName -> FuncInfo
var interfaceMapping map[string]map[string]*info.Func        // pkg -> interfaceName -> FuncInfo
var typeMethodMapping map[reflect.Type]map[string]*info.Func // reflect.Type -> interfaceName -> FuncInfo

func init() {
	funcPCMapping = make(map[uintptr]*info.Func)
	funcInfoMapping = make(map[string]map[string]*info.Func)
	funcFullNameMapping = make(map[string]*info.Func)
	interfaceMapping = make(map[string]map[string]*info.Func)
	varAddrMapping = make(map[uintptr]*info.Func)

	// this will consume all staged func infos in runtime,
	// and set registerFuncInfo for later registering
	info.SetupRegisterHandler(RegisterFunc)
}

func RegisterFunc(funcInfo *info.Func) {
	if funcInfo == nil {
		panic("funcInfo is nil")
	}
	funcInfos = append(funcInfos, funcInfo)
	registerIndex(funcInfo)
}

func GetFuncs() []*info.Func {
	return funcInfos
}

func InfoFunc(fn interface{}) *info.Func {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("given type is not a func: %T", fn))
	}
	// deref to pc
	pc := v.Pointer()
	return funcPCMapping[pc]
}

func InfoVar(addr interface{}) *info.Func {
	v := reflect.ValueOf(addr)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("given type is not a pointer: %T", addr))
	}
	ptr := v.Pointer()
	return varAddrMapping[ptr]
}

// maybe rename to FuncForPC
func InfoPC(pc uintptr) *info.Func {
	return funcPCMapping[pc]
}

func InfoVarAddr(addr uintptr) *info.Func {
	return varAddrMapping[addr]
}

// maybe rename to FuncForGeneric
func Info(pkg string, identityName string) *info.Func {
	return funcInfoMapping[pkg][identityName]
}

// GetFuncByPkg:
//
//	pkg.Func
//	pkg.Recv.Func
//	pkg.(*Recv).Func
func GetFuncByPkg(pkg string, name string) *info.Func {
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

func GetFuncByFullName(fullName string) *info.Func {
	f := funcFullNameMapping[fullName]
	if f != nil {
		return f
	}
	return getInterfaceOrGenericByFullName(fullName)
}

func GetTypeMethods(typ reflect.Type) map[string]*info.Func {
	return getTypeMethodMapping()[typ]
}

func getInterfaceOrGenericByFullName(fullName string) *info.Func {
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

// Deprecated: now xgo can make package automatically depends
// on this package(github.com/xhd2015/xgo/runtime/functab),
// so we don't need to do struct inspection here anymore.
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
	var infoVar interface{}
	if varField.IsValid() {
		infoVar = varField.Interface()
	}
	info := &core.FuncInfo{
		Kind:         fnKind,
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
		Func: f, // nil for geneirc
		Var:  infoVar,

		RecvName: recvName,
		ArgNames: argNames,
		ResNames: resNames,
	}
	funcInfos = append(funcInfos, info)
	registerIndex(info)
}

func registerIndex(funcInfo *info.Func) {
	identityName := funcInfo.IdentityName
	generic := funcInfo.Generic
	interface_ := funcInfo.Interface
	recvTypeName := funcInfo.RecvType
	pkgPath := funcInfo.Pkg
	fnKind := funcInfo.Kind
	infoVar := funcInfo.Var

	closure := funcInfo.Closure

	f := funcInfo.Func

	var pcFullName string
	var firstArgCtx bool
	var lastResErr bool
	var pc uintptr
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
			pcFullName = runtime.XgoGetFullPCName(pc)
		} else {
			if (closure || fnKind == core.Kind_Var || fnKind == core.Kind_VarPtr || fnKind == core.Kind_Const) && identityName != "" {
				pcFullName = pkgPath + "." + identityName
			}
		}
	}
	funcInfo.PC = pc

	if funcInfo.FullName == "" {
		funcInfo.FullName = pcFullName
	} else if funcInfo.FullName != pcFullName {
		panic(fmt.Errorf("func name mismatch: %s != %s", funcInfo.FullName, pcFullName))
	}

	// brief info
	funcInfo.FirstArgCtx = firstArgCtx
	funcInfo.LastResultErr = lastResErr

	// register index
	if funcInfo.FullName != "" {
		funcFullNameMapping[funcInfo.FullName] = funcInfo
	}

	if !generic && funcInfo.PC != 0 {
		funcPCMapping[funcInfo.PC] = funcInfo
	}
	if identityName != "" {
		pkgMapping := funcInfoMapping[pkgPath]
		if pkgMapping == nil {
			pkgMapping = make(map[string]*info.Func, 1)
			funcInfoMapping[pkgPath] = pkgMapping
		}
		pkgMapping[identityName] = funcInfo
	}
	if interface_ && recvTypeName != "" {
		pkgMapping := interfaceMapping[pkgPath]
		if pkgMapping == nil {
			pkgMapping = make(map[string]*info.Func, 1)
			interfaceMapping[pkgPath] = pkgMapping
		}
		pkgMapping[recvTypeName] = funcInfo
	}
	if fnKind == core.Kind_Var {
		if infoVar != nil {
			// infoVar is &v
			varAddr := reflect.ValueOf(infoVar).Pointer()
			varAddrMapping[varAddr] = funcInfo
		}
	}
}

var mappingTypeOnce sync.Once

func getTypeMethodMapping() map[reflect.Type]map[string]*info.Func {
	mappingTypeOnce.Do(initTypeMethodMapping)
	return typeMethodMapping
}

func initTypeMethodMapping() {
	typeMethodMapping = make(map[reflect.Type]map[string]*info.Func)
	for _, funcInfo := range funcInfos {
		registerTypeMethod(funcInfo)
	}
}

func registerTypeMethod(funcInfo *info.Func) {
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
		methodMapping = make(map[string]*info.Func, 1)
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
