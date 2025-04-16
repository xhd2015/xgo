package resolve

import (
	"go/ast"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

func (c *Scope) recordTrap(sel *ast.SelectorExpr) bool {
	typInfo := c.tryResolvePkgRef(sel)
	if types.IsUnknown(typInfo) {
		return false
	}
	pkgObject, ok := typInfo.(types.PkgVariable)
	if !ok {
		return false
	}
	pkgPath := pkgObject.PkgPath
	name := pkgObject.Name
	if name == "Patch" || name == "Mock" {
		return pkgPath == constants.RUNTIME_MOCK_PKG
	}
	if name == "AddFuncInterceptor" || name == "MarkIntercept" {
		return pkgPath == constants.RUNTIME_TRAP_PKG
	}
	if name == "Record" || name == "RecordCall" || name == "RecordResult" {
		return pkgPath == constants.RUNTIME_TRACE_PKG
	}
	if name == "InfoFunc" || name == "InfoVar" {
		return pkgPath == constants.RUNTIME_FUNCTAB_PKG
	}
	if name == "AddInterceptor" || name == "MarkInterceptAll" {
		if pkgPath == constants.RUNTIME_TRAP_PKG {
			c.Global.Recorder.HasTrapInterceptorRef = true
		}
		return false
	}
	return false
}

// `mock.Patch(fn,...)`
// fn examples:
//
//	(*ipc.Reader).Next
//	vr.Reader.Next where vr.Reader is an embedded field: struct { *ipc.Reader }
func (c *Scope) recordMockRef(fn ast.Expr) {
	fnInfo := c.resolveInfo(fn)
	if types.IsUnknown(fnInfo) {
		return
	}
	recorder := c.Global.Recorder

	var pkgPath string
	var name string
	var field string
	var isVar bool
	switch fnInfo := fnInfo.(type) {
	case types.PkgTypeMethod:
		if namedType, ok := fnInfo.Recv.(types.NamedType); ok {
			pkgPath = namedType.PkgPath
			name = namedType.Name
			field = fnInfo.Name
		}
	case types.PkgFunc:
		pkgPath = fnInfo.PkgPath
		name = fnInfo.Name
	case types.Pointer:
		// pointer to types.PackageVariable
		if pkgVar, ok := fnInfo.Value.(types.PkgVariable); ok {
			pkgPath = pkgVar.PkgPath
			name = pkgVar.Name
			isVar = true
		}
	default:
		return
	}
	if pkgPath == "" || name == "" {
		return
	}
	if _, allow := config.CheckInstrument(pkgPath); !allow {
		return
	}

	topRecord := recorder.GetOrInit(pkgPath).GetOrInit(name)
	if isVar {
		topRecord.HasVarTrap = true
	} else {
		if field == "" {
			topRecord.HasMockRef = true
		} else {
			topRecord.AddMockName(field)
		}
	}
}

// try simply resolve as pkg.Name, to detect for mock.Patch calls
func (c *Scope) tryResolvePkgRef(sel *ast.SelectorExpr) types.Info {
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return types.Unknown{}
	}
	if c.Has(idt.Name) {
		return types.Unknown{}
	}
	decl := c.Package.Decls[idt.Name]
	if decl != nil {
		return types.Unknown{}
	}
	imp, ok := c.File.Imports[idt.Name]
	if !ok {
		return types.Unknown{}
	}
	return types.PkgVariable{
		PkgPath: imp,
		Name:    sel.Sel.Name,
	}
}
