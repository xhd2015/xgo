package functab

import (
	"strings"
	"sync"
)

const __XGO_SKIP_TRAP = true

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_for_each_func(f func(pkgName string, funcName string, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string)) {
	panic("xgo failed to link __xgo_link_for_each_func2")
}

type FuncInfo struct {
	FullName string
	Name     string
	PkgPath  string
	RecvType string
	RecvPtr  string

	PC       uintptr
	Func     interface{}
	RecvName string
	ArgNames []string
	ResNames []string
}

var funcInfos []*FuncInfo
var funcInfoMapping map[string]*FuncInfo

func GetFuncs() []*FuncInfo {
	ensureMapping()
	return funcInfos
}

func GetFunc(fullName string) *FuncInfo {
	ensureMapping()
	return funcInfoMapping[fullName]
}

func GetFuncByPkg(pkgPath string, name string) *FuncInfo {
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
		funcInfoMapping = map[string]*FuncInfo{}
		__xgo_link_for_each_func(func(pkgName string, funcName string, pc uintptr, fn interface{}, recvName string, argNames, resNames []string) {
			name := strings.TrimPrefix(funcName, pkgName+".")
			info := &FuncInfo{
				FullName: funcName,
				Name:     name,
				PC:       pc,
				Func:     fn,
				RecvName: recvName,
				ArgNames: argNames,
				ResNames: resNames,
			}
			funcInfos = append(funcInfos, info)
			funcInfoMapping[info.FullName] = info
		})
	})
}
