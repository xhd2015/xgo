package functab

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const __XGO_SKIP_TRAP = true

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_for_each_func(f func(pkgName string, funcName string, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string)) {
	panic("failed to link __xgo_link_for_each_func")
}

type FuncInfo struct {
	FullName string
	Name     string
	Pkg      string
	RecvType string
	RecvPtr  bool

	PC       uintptr     `json:"-"`
	Func     interface{} `json:"-"`
	RecvName string
	ArgNames []string
	ResNames []string
}

func (c *FuncInfo) DisplayName() string {
	if c.RecvType != "" {
		return c.RecvType + "." + c.Name
	}
	return c.Name
}

var funcInfos []*FuncInfo
var funcInfoMapping map[string]*FuncInfo
var funcPCMapping map[uintptr]*FuncInfo // pc->FuncInfo

func GetFuncs() []*FuncInfo {
	ensureMapping()
	return funcInfos
}

func GetFunc(fullName string) *FuncInfo {
	ensureMapping()
	return funcInfoMapping[fullName]
}

func Info(fn interface{}) *FuncInfo {
	ensureMapping()
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("given type is not a func: %T", fn))
	}
	// deref to pc
	pc := v.Pointer()
	return funcPCMapping[pc]
}

func InfoPC(pc uintptr) *FuncInfo {
	ensureMapping()
	return funcPCMapping[pc]
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
		funcPCMapping = make(map[uintptr]*FuncInfo)
		__xgo_link_for_each_func(func(pkgPath string, funcName string, pc uintptr, fn interface{}, recvName string, argNames, resNames []string) {
			// prefix := pkgPath + "."
			_, recvTypeName, recvPtr, name := ParseFuncName(funcName[len(pkgPath)+1:], false)
			info := &FuncInfo{
				FullName: funcName,
				Name:     name,
				Pkg:      pkgPath,
				RecvType: recvTypeName,
				RecvPtr:  recvPtr,

				//
				PC:       pc,
				Func:     fn,
				RecvName: recvName,
				ArgNames: argNames,
				ResNames: resNames,
			}
			funcInfos = append(funcInfos, info)
			funcInfoMapping[info.FullName] = info
			funcPCMapping[info.PC] = info
		})
	})
}

// a/b/c.A
// a/b/c.(*C).X
// a/b/c.C.Y
// a/b/c.Z
func ParseFuncName(fullName string, hasPkg bool) (pkgPath string, recvName string, recvPtr bool, funcName string) {
	s := fullName
	funcNameDot := strings.LastIndex(s, ".")
	if funcNameDot < 0 {
		funcName = s
		return
	}
	funcName = s[funcNameDot+1:]
	s = s[:funcNameDot]

	recvName = s
	if hasPkg {
		recvDot := strings.LastIndex(s, ".")
		if recvDot < 0 {
			pkgPath = s
			return
		}
		recvName = s[recvDot+1:]
		s = s[:recvDot]
	}

	recvName = strings.TrimPrefix(recvName, "(")
	recvName = strings.TrimSuffix(recvName, ")")
	if strings.HasPrefix(recvName, "*") {
		recvPtr = true
		recvName = recvName[1:]
	}
	pkgPath = s

	return
}
