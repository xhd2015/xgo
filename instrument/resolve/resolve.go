package resolve

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
)

func (c *Scope) requireTrap(sel *ast.SelectorExpr) bool {
	typInfo, ok := c.tryResolvePkgRef(sel)
	if !ok {
		return false
	}
	if typInfo.Name != "" {
		return false
	}
	name := typInfo.TopName
	pkgPath := typInfo.PkgPath
	if name == "Patch" || name == "Mock" {
		return pkgPath == constants.RUNTIME_MOCK_PKG
	}
	if name == "AddFuncInterceptor" {
		return pkgPath == constants.RUNTIME_TRAP_PKG
	}
	if name == "Record" || name == "RecordCall" || name == "RecordResult" {
		return pkgPath == constants.RUNTIME_TRACE_PKG
	}
	return false
}

// `mock.Patch(fn,...)`
// fn examples:
//
//	(*ipc.Reader).Next
//	vr.Reader.Next where vr.Reader is an embedded field: struct { *ipc.Reader }
func (c *Scope) resolveMockRef(fn ast.Expr) {
	typeInfo, ok := c.resolveType(fn)
	if !ok {
		return
	}
	recorder := c.Global.Recorder
	pkgPath := typeInfo.PkgPath
	topName := typeInfo.TopName
	name := typeInfo.Name

	topRecord := recorder.GetOrInit(pkgPath).GetOrInit(topName)
	if name == "" {
		topRecord.HasMockPatch = true
	} else {
		topRecord.AddMockName(name)
	}
}

type TypeInfo struct {
	PkgPath string
	TopName string
	Name    string
}

func (c *Scope) resolveType(expr ast.Expr) (TypeInfo, bool) {
	if expr == nil {
		return TypeInfo{}, false
	}
	switch expr := expr.(type) {
	case *ast.Ident:
		def := c.GetDef(expr.Name)
		if def == nil {
			return TypeInfo{}, false
		}
		// def example:
		//   &flight.Writer{}
		if def.Index == -1 {
			// TODO
		}
	case *ast.SelectorExpr:
		selName := expr.Sel.Name
		if isBlankName(selName) {
			return TypeInfo{}, false
		}
		if idt, idtOK := expr.X.(*ast.Ident); idtOK {
			pkgPath, ok := c.Global.Imports[idt.Name]
			if ok {
				// pkg.Name
				return TypeInfo{
					PkgPath: pkgPath,
					TopName: selName,
				}, true
			}
			def := c.GetDef(idt.Name)
			if def != nil {
				// var.Name
				return c.resolveDefType(def)
			}
			return TypeInfo{}, false
		}
		typeInfo, ok := c.tryResolvePkgType(expr.X)
		if ok {
			// (*ipc.Reader).Next or ipc.Reader.Next
			return typeInfo, true
		}
		// var.Field.Func
		typeInfo, ok = c.tryResolveVarDotField(expr.X)
		if ok {
			if typeInfo.Name == "" {
				return TypeInfo{
					PkgPath: typeInfo.PkgPath,
					TopName: typeInfo.TopName,
					Name:    selName,
				}, true
			}
			return typeInfo, true
		}
	case *ast.StarExpr:
		return c.resolveType(expr.X)
	}
	return TypeInfo{}, false
}

// pkg.Name
func (c *Scope) tryResolvePkgRef(sel *ast.SelectorExpr) (typeInfo TypeInfo, ok bool) {
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return TypeInfo{}, false
	}
	imp, ok := c.Global.Imports[idt.Name]
	if !ok {
		return TypeInfo{}, false
	}
	return TypeInfo{
		PkgPath: imp,
		TopName: sel.Sel.Name,
	}, true
}

// &flight.Writer{}
func (c *Scope) tryResolveCompositeLit(expr ast.Expr) (typeInfo TypeInfo, ok bool) {
	unary, ok := expr.(*ast.UnaryExpr)
	if ok {
		if unary.Op != token.AND {
			return TypeInfo{}, false
		}
		expr = unary.X
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return TypeInfo{}, false
	}
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok {
		return TypeInfo{}, false
	}
	return c.tryResolvePkgRef(sel)
}

// (*ipc.Reader).Next or ipc.Reader.Next
func (c *Scope) tryResolvePkgType(expr ast.Expr) (typeInfo TypeInfo, ok bool) {
	paren, ok := expr.(*ast.ParenExpr)
	if ok {
		starExpr, ok := paren.X.(*ast.StarExpr)
		if !ok {
			return TypeInfo{}, false
		}
		expr = starExpr.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return TypeInfo{}, false
	}
	return c.tryResolvePkgRef(sel)
}

func (c *Scope) tryResolveVarDotField(expr ast.Expr) (typeInfo TypeInfo, ok bool) {
	// var.Field
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return TypeInfo{}, false
	}
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return TypeInfo{}, false
	}
	def := c.GetDef(idt.Name)
	if def == nil {
		return TypeInfo{}, false
	}

	typeInfo, ok = c.resolveDefType(def)
	if !ok {
		return TypeInfo{}, false
	}
	pkgPath := typeInfo.PkgPath
	topName := typeInfo.TopName
	// could be a struct
	if typeInfo.Name == "" {
		structFieldType, ok := c.tryGetStructFieldType(pkgPath, topName, sel.Sel.Name)
		if ok {
			return TypeInfo{
				PkgPath: structFieldType.PkgPath,
				TopName: structFieldType.TopName,
			}, true
		}
	}

	return TypeInfo{}, false
}
func (c *Scope) tryGetStructFieldType(pkgPath string, declName string, refFieldName string) (typeInfo TypeInfo, ok bool) {
	// check if pkgPath.name is a struct
	pkg, err := c.Global.Packages.Load(pkgPath)
	if err != nil {
		panic(fmt.Sprintf("resolveDefType: %s %s", pkgPath, err))
	}
	if pkg == nil {
		return TypeInfo{}, false
	}

	Collect(&edit.Packages{
		Packages: []*edit.Package{pkg},
	}, Options{
		RecordTypes: true,
	})
	decl := pkg.Decls[declName]
	if decl == nil {
		return TypeInfo{}, false
	}
	structType, ok := decl.Type.(*ast.StructType)
	if !ok {
		return TypeInfo{}, false
	}
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		return TypeInfo{}, false
	}
	var foundField *ast.Field
	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			for _, fieldName := range field.Names {
				if fieldName.Name == refFieldName {
					foundField = field
					break
				}
			}
		} else if field.Type != nil {
			typName := getEmbedFieldName(field.Type)
			if typName == refFieldName {
				foundField = field
				break
			}
		}
	}
	imports := getFileImports(decl.File.File.Syntax)
	pkgCtx := &Scope{
		Global: &GlobalScope{
			Packages:      c.Global.Packages,
			Package:       pkg,
			File:          decl.File,
			PkgScopeNames: pkg.Decls,
			Imports:       imports,
			Recorder:      c.Global.Recorder,
			detectVarTrap: c.Global.detectVarTrap,
			detectMock:    c.Global.detectMock,
		},
	}
	return pkgCtx.resolveType(foundField.Type)
}

func getEmbedFieldName(typ ast.Expr) string {
	starExpr, ok := typ.(*ast.StarExpr)
	if ok {
		typ = starExpr.X
	}
	sel, ok := typ.(*ast.SelectorExpr)
	if ok {
		return sel.Sel.Name
	}
	ident, ok := typ.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}
func (c *Scope) resolveDefType(def *Define) (typeInfo TypeInfo, ok bool) {
	if def.Index == -1 {
		// var := pkg.Name{}
		typeInfo, ok = c.tryResolveCompositeLit(def.Expr)
		if ok {
			return typeInfo, true
		}
	}
	// not supported yet
	return TypeInfo{}, false
}
