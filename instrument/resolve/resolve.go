package resolve

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// this package resolves type of an AST Expr.
// a type is defined as giving the following information:
// - package path
// - type name
// - has a receivier
// - is the receiver a pointer
// - is the receiver generic
//
// the result type are split into 3 parts:
// - package-level variable
// - package-level function
// - package-level method of a type

// TODO: can we support generic functions and methods even in go1.18 and go1.19 with the
// help of resolver?
// through a rewriting of mock.Patch? If we know the function has been instrumented,
// we record that to some place, and replace the call with mock.AutoGenPatchGeneric(which is generated on the fly), in that function the pc lookup is done by inspecting directly.
// or more generally, we generate the

func (c *Scope) requireTrap(sel *ast.SelectorExpr) bool {
	typInfo := c.tryResolvePkgRef(sel)
	if types.IsUnknown(typInfo) {
		return false
	}
	pkgObject, ok := typInfo.(types.Variable)
	if !ok {
		return false
	}
	pkgPath := pkgObject.PkgPath
	name := pkgObject.Name
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
func (c *Scope) recordMockRef(fn ast.Expr) {
	typeInfo := c.resolveType(fn)
	if types.IsUnknown(typeInfo) {
		return
	}
	recorder := c.Global.Recorder

	var pkgPath string
	var name string
	var field string
	switch typeInfo := typeInfo.(type) {
	case types.Variable:
		pkgPath = typeInfo.PkgPath
		name = typeInfo.Name
	case types.VariableField:
		pkgPath = typeInfo.Variable.PkgPath
		name = typeInfo.Variable.Name
		field = typeInfo.Field
	default:
		return
	}

	topRecord := recorder.GetOrInit(pkgPath).GetOrInit(name)
	if field == "" {
		topRecord.HasMockRef = true
	} else {
		topRecord.AddMockName(field)
	}
}

// a -> look for definition
// a.b -> look for a's definition first, a could a package name, a type or a variable.
// a.b.c -> look for a.b's definition

func (c *Scope) resolveType(expr ast.Expr) types.Type {
	if expr == nil {
		return types.Unknown{}
	}
	typeInfo, ok := c.Global.ObjectInfo[expr]
	if ok {
		return typeInfo
	}
	typeInfo = c.doResolveType(expr)
	if c.Global.ObjectInfo == nil {
		c.Global.ObjectInfo = make(map[ast.Expr]types.Type, 1)
	}
	c.Global.ObjectInfo[expr] = typeInfo
	return typeInfo
}

func (c *Scope) doResolveType(expr ast.Expr) types.Type {
	switch expr := expr.(type) {
	case *ast.Ident:
		name := expr.Name
		def := c.GetDef(name)
		if def != nil {
			return c.resolveDefType(def)
		}
		if !c.Has(expr.Name) {
			// check decl before imports
			decl := c.Global.PkgScopeNames[name]
			if decl != nil {
				if decl.Kind == edit.DeclKindType {
					namedType := types.NamedType{
						PkgPath: c.Global.Package.LoadPackage.GoPackage.ImportPath,
						Name:    name,
					}
					if c.Global.NamedTypeToDecl == nil {
						c.Global.NamedTypeToDecl = make(map[types.NamedType]*edit.Decl, 1)
					}
					c.Global.NamedTypeToDecl[namedType] = decl
					return namedType
				} else if decl.Kind == edit.DeclKindVar {
					var varType types.Type
					if decl.Type != nil {
						varType = c.resolveType(decl.Type)
					} else if decl.Value != nil {
						varType = c.resolveType(decl.Value)
					}
					return types.Variable{
						PkgPath: c.Global.Package.LoadPackage.GoPackage.ImportPath,
						Name:    name,
						Type:    varType,
					}
				}
				return types.Unknown{}
			}
			pkgPath, ok := c.Global.Imports[expr.Name]
			if ok {
				// this only refers to a package
				return types.ImportPath(pkgPath)
			}
		}
	case *ast.SelectorExpr:
		selName := expr.Sel.Name
		if xInfo := c.resolveType(expr.X); !types.IsUnknown(xInfo) {
			// if xType resolves to a Decl,
			// check if Name resolves to method of that Decl
			if xVar, ok := xInfo.(types.Variable); ok {
				if namedType, ok := xVar.Type.(types.NamedType); ok {
					decl := c.Global.NamedTypeToDecl[namedType]
					if decl != nil {
						method := decl.Methods[selName]
						if method != nil && method.RecvPtr {
							// update expr.X's type to be ptr
							xVar.Type = types.Ptr{
								Elem: xVar.Type,
							}
							c.Global.ObjectInfo[expr.X] = xVar
						}
						return types.Func{
							Recv: namedType,
							// TODO params
						}
					}
				}
			}
		}
		// X.Y X has a type
		if isBlankName(selName) {
			return types.Unknown{}
		}
		if idt, idtOK := expr.X.(*ast.Ident); idtOK {
			pkgPath, ok := c.Global.Imports[idt.Name]
			if ok {
				// pkg.Name
				return types.Variable{
					PkgPath: pkgPath,
					Name:    selName,
				}
			}
			def := c.GetDef(idt.Name)
			if def != nil {
				// var.Name
				return c.resolveDefType(def)
			}
			return types.Unknown{}
		}
		typeInfo := c.tryResolvePkgType(expr.X)
		if !types.IsUnknown(typeInfo) {
			// (*ipc.Reader).Next or ipc.Reader.Next
			return typeInfo
		}
		// var.Field.Func
		typeInfo = c.tryResolveVarDotField(expr.X)
		if !types.IsUnknown(typeInfo) {
			pkgObj, ok := typeInfo.(types.Variable)
			if ok {
				return types.VariableField{
					Variable: pkgObj,
					Field:    selName,
				}
			}
			panic(fmt.Sprintf("TODO unhandled type: %T", typeInfo))
		}
	case *ast.StarExpr:
		return c.resolveType(expr.X)
	case *ast.UnaryExpr:
		// TODO
		// &var
		// could be a variable
	case *ast.BasicLit:
		return getBasicLitConstType(expr.Kind)
	}
	return types.Unknown{}
}

func getBasicLitConstType(kind token.Token) types.Type {
	switch kind {
	case token.INT:
		return types.Basic("int")
	case token.STRING:
		return types.Basic("string")
	case token.CHAR:
		return types.Basic("byte")
	case token.FLOAT:
		return types.Basic("float64")
	}
	return types.Unknown{}
}

// pkg.Name
func (c *Scope) tryResolvePkgRef(sel *ast.SelectorExpr) types.Type {
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return types.Unknown{}
	}
	imp, ok := c.Global.Imports[idt.Name]
	if !ok {
		return types.Unknown{}
	}
	return types.Variable{
		PkgPath: imp,
		Name:    sel.Sel.Name,
	}
}

// &flight.Writer{}
func (c *Scope) tryResolveCompositeLit(expr ast.Expr) types.Type {
	unary, ok := expr.(*ast.UnaryExpr)
	if ok {
		if unary.Op != token.AND {
			return types.Unknown{}
		}
		expr = unary.X
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return types.Unknown{}
	}
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok {
		return types.Unknown{}
	}
	return c.tryResolvePkgRef(sel)
}

// (*ipc.Reader).Next or ipc.Reader.Next
func (c *Scope) tryResolvePkgType(expr ast.Expr) types.Type {
	paren, ok := expr.(*ast.ParenExpr)
	if ok {
		starExpr, ok := paren.X.(*ast.StarExpr)
		if !ok {
			return types.Unknown{}
		}
		expr = starExpr.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return types.Unknown{}
	}
	return c.tryResolvePkgRef(sel)
}

func (c *Scope) tryResolveVarDotField(expr ast.Expr) types.Type {
	// var.Field
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return types.Unknown{}
	}
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return types.Unknown{}
	}
	def := c.GetDef(idt.Name)
	if def == nil {
		return types.Unknown{}
	}

	typeInfo := c.resolveDefType(def)
	if types.IsUnknown(typeInfo) {
		return types.Unknown{}
	}
	if typeInfo, ok := typeInfo.(types.Variable); ok {
		// could be a struct
		pkgPath := typeInfo.PkgPath
		typeName := typeInfo.Name
		structFieldType := c.tryGetStructFieldType(pkgPath, typeName, sel.Sel.Name)
		if !types.IsUnknown(structFieldType) {
			return structFieldType
		}
	}

	return types.Unknown{}
}
func (c *Scope) tryGetStructFieldType(pkgPath string, declName string, refFieldName string) types.Type {
	// check if pkgPath.name is a struct
	pkg, _, err := c.Global.Packages.Load(pkgPath)
	if err != nil {
		panic(fmt.Sprintf("resolveDefType: %s %s", pkgPath, err))
	}
	if pkg == nil {
		return types.Unknown{}
	}

	Collect(&edit.Packages{
		Packages: []*edit.Package{pkg},
	})

	decl := pkg.Decls[declName]
	if decl == nil {
		return types.Unknown{}
	}
	structType, ok := decl.Type.(*ast.StructType)
	if !ok {
		return types.Unknown{}
	}
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		return types.Unknown{}
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
func (c *Scope) resolveDefType(def *Define) types.Type {
	if def.Index == -1 {
		// var := pkg.Name{}
		typeInfo := c.tryResolveCompositeLit(def.Expr)
		if !types.IsUnknown(typeInfo) {
			return typeInfo
		}
	}
	// not supported yet
	return types.Unknown{}
}
