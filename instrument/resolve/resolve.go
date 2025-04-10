package resolve

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"

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
	fnInfo := c.resolveInfo(fn)
	if types.IsUnknown(fnInfo) {
		return
	}
	recorder := c.Global.Recorder

	var pkgPath string
	var name string
	var field string
	switch typeInfo := fnInfo.(type) {
	case types.Method:
		if namedType, ok := typeInfo.Recv.(types.NamedType); ok {
			pkgPath = namedType.PkgPath
			name = namedType.Name
			field = typeInfo.Name
		}
	case types.Func:
		pkgPath = typeInfo.PkgPath
		name = typeInfo.Name
	default:
		return
	}
	if pkgPath == "" || name == "" {
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
func (c *Scope) resolveInfo(expr ast.Expr) types.Info {
	if expr == nil {
		return types.Unknown{}
	}
	typeInfo, ok := c.Global.ExprInfo[expr]
	if ok {
		return typeInfo
	}
	typeInfo = c.doResolveInfo(expr)
	if c.Global.ExprInfo == nil {
		c.Global.ExprInfo = make(map[ast.Expr]types.Info, 1)
	}
	c.Global.ExprInfo[expr] = typeInfo
	return typeInfo
}

func (c *Scope) resolveType(expr ast.Expr) types.Type {
	info := c.resolveInfo(expr)
	if types.IsUnknown(info) {
		return types.Unknown{}
	}
	typ, ok := info.(types.Type)
	if ok {
		return typ
	}
	return types.Unknown{}
}

func (c *Scope) doResolveInfo(expr ast.Expr) types.Info {
	switch expr := expr.(type) {
	case *ast.Ident:
		name := expr.Name
		def := c.GetDef(name)
		if def != nil {
			return c.resolveDefType(def)
		}
		if !c.Has(expr.Name) {
			// check decl before imports
			decl := c.Package.Decls[name]
			if decl != nil {
				if decl.Kind == edit.DeclKindType {
					typ := c.resolveType(decl.Type)
					if types.IsUnknown(typ) {
						return types.Unknown{}
					}
					pkgPath := c.Package.PkgPath()
					namedType := types.NamedType{
						PkgPath: pkgPath,
						Name:    name,
						Type:    typ,
					}
					c.recordPkgNameDecl(pkgPath, name, decl)
					return namedType
				} else if decl.Kind == edit.DeclKindVar {
					var varType types.Type
					if decl.Type != nil {
						varType = c.resolveType(decl.Type)
					} else if decl.Value != nil {
						varType = c.resolveType(decl.Value)
					}
					return types.Variable{
						PkgPath: c.Package.PkgPath(),
						Name:    name,
						Type_:   varType,
					}
				} else if decl.Kind == edit.DeclKindFunc {
					return types.Func{
						PkgPath: c.Package.PkgPath(),
						Name:    name,
						// TODO params
					}
				}
				return types.Unknown{}
			}
			pkgPath, ok := c.File.Imports[expr.Name]
			if ok {
				// this only refers to a package
				return types.ImportPath(pkgPath)
			}
		}
	case *ast.SelectorExpr:
		selName := expr.Sel.Name
		xInfo := c.resolveInfo(expr.X)
		if types.IsUnknown(xInfo) {
			return types.Unknown{}
		}

		// if xType resolves to a Decl,
		// check if Name resolves to method of that Decl
		var valType types.Type
		if obj, ok := xInfo.(types.Object); ok {
			valType = obj.Type()
		}
		if valType != nil && !types.IsUnknown(valType) {
			var isPtr bool
			resolveType := valType
			if ptrType, ok := valType.(types.Ptr); ok {
				// *T
				isPtr = true
				resolveType = ptrType.Elem
			}
			if namedType, ok := resolveType.(types.NamedType); ok {
				decl := c.getPkgNameDecl(namedType.PkgPath, namedType.Name)
				var method *edit.FuncDecl
				if decl != nil {
					method = decl.Methods[selName]
				}
				if method != nil {
					if !isPtr && method.RecvPtr {
						// update expr.X's type to be ptr
						c.Global.ExprInfo[expr.X] = takeAddr(xInfo)
					}
					return types.Method{
						Name: selName,
						Recv: namedType,
						// TODO params
					}
				} else {
					// check fields
					ud, ok := valType.Underlying().(types.Struct)
					if ok {
						var fieldType types.Type
						for _, field := range ud.Fields {
							if field.Name == selName {
								fieldType = field.Type
								break
							}
						}
						if fieldType != nil && !types.IsUnknown(fieldType) {
							return types.Value{
								Type_: fieldType,
							}
						}
					}
				}
			}
			return types.Unknown{}
		}

		impPath, ok := xInfo.(types.ImportPath)
		if ok {
			// pkg.Name
			pkgPath := string(impPath)
			pkg := c.loadPackage(pkgPath)
			if pkg == nil {
				return types.Unknown{}
			}
			if decl := pkg.Decls[selName]; decl != nil {
				pkgScope := c.newFileScope(pkg, decl.File)
				return pkgScope.resolveInfo(decl.Ident)
			}
			return types.Unknown{}
		}

		if xInfo := c.resolveInfo(expr.X); !types.IsUnknown(xInfo) {

		}
		// X.Y X has a type
		if isBlankName(selName) {
			return types.Unknown{}
		}
		if idt, idtOK := expr.X.(*ast.Ident); idtOK {
			pkgPath, ok := c.File.Imports[idt.Name]
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
				// TODO
				_ = pkgObj
				return types.Unknown{}
				// return types.VariableField{
				// 	Variable: pkgObj,
				// 	Field:    selName,
				// }
			}
			panic(fmt.Sprintf("TODO unhandled type: %T", typeInfo))
		}
	case *ast.StarExpr:
		xInfo := c.resolveInfo(expr.X)
		if types.IsUnknown(xInfo) {
			return types.Unknown{}
		}
		if typ, ok := xInfo.(types.Type); ok {
			return types.Ptr{Elem: typ}
		}
		if obj, ok := xInfo.(types.Object); ok {
			derefType := deref(obj.Type())
			if types.IsUnknown(derefType) {
				return types.Unknown{}
			}
			return types.Value{Type_: derefType}
		}
	case *ast.UnaryExpr:
		// TODO
		// &var
		// could be a variable
		vr := c.resolveInfo(expr.X)
		if types.IsUnknown(vr) {
			return types.Unknown{}
		}
		switch vr := vr.(type) {
		case types.Variable:
			if expr.Op == token.AND {
				return types.Variable{
					PkgPath: vr.PkgPath,
					Name:    vr.Name,
					Type_:   types.Ptr{Elem: vr.Type_},
				}
			}
		case types.Literal:
			if expr.Op == token.AND {
				return types.Literal{
					Type_: types.Ptr{Elem: vr.Type_},
				}
			}
		}
		return types.Unknown{}
	case *ast.CompositeLit:
		// generic
		basicType := expr.Type
		if idxExpr, ok := expr.Type.(*ast.IndexExpr); ok {
			basicType = idxExpr.X
		}

		typ := c.resolveType(basicType)
		if types.IsUnknown(typ) {
			return types.Unknown{}
		}
		return types.Literal{
			Type_: typ,
		}
	case *ast.IndexExpr:
		info := c.resolveInfo(expr.X)
		// if info is a Type, then this is
		// generic type
		if typ, ok := info.(types.Type); ok {
			return typ
		}
		// otherwise this yields an element type
		// from a slice or array
		// TODO: support array
		return types.Unknown{}
	case *ast.BasicLit:
		return getBasicLitConstType(expr.Kind)
	case *ast.StructType:
		fields := make([]types.StructField, len(expr.Fields.List))
		for i, field := range expr.Fields.List {
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					typ := c.resolveInfo(field.Type)
					if typ, ok := typ.(types.Type); ok {
						fields[i] = types.StructField{
							Name: name.Name,
							Type: typ,
						}
					}
				}
			} else {
				typ := c.resolveInfo(field.Type)
				if typ, ok := typ.(types.Type); ok {
					fields[i] = types.StructField{
						Name: getEmbedFieldName(field.Type),
						Type: typ,
					}
				}
			}
		}
		return types.Struct{
			Fields: fields,
		}
	}
	return types.Unknown{}
}

func (c *Scope) recordPkgNameDecl(pkgPath string, name string, decl *edit.Decl) {
	if c.Global.NamedTypeToDecl == nil {
		c.Global.NamedTypeToDecl = make(map[PkgName]*edit.Decl, 1)
	}
	c.Global.NamedTypeToDecl[PkgName{
		PkgPath: pkgPath,
		Name:    name,
	}] = decl
}

func (c *Scope) getPkgNameDecl(pkgPath string, name string) *edit.Decl {
	return c.Global.NamedTypeToDecl[PkgName{
		PkgPath: pkgPath,
		Name:    name,
	}]
}

func (c *Scope) loadPackage(pkgPath string) *edit.Package {
	pkg, _, err := c.Global.Packages.Load(pkgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %s\n", pkgPath, err)
		return nil
	}
	if pkg == nil {
		return nil
	}
	Collect(&edit.Packages{
		Packages: []*edit.Package{pkg},
	})
	return pkg
}

func deref(typ types.Info) types.Type {
	if ptr, ok := typ.(types.Ptr); ok {
		return ptr.Elem
	}
	return types.Unknown{}
}

func takeAddr(info types.Info) types.Info {
	switch info := info.(type) {
	case types.Variable:
		return types.Variable{
			PkgPath: info.PkgPath,
			Name:    info.Name,
			Type_:   types.Ptr{Elem: info.Type_},
		}
	case types.Literal:
		return types.Literal{
			Type_: types.Ptr{Elem: info.Type_},
		}
	case types.Value:
		return types.Value{
			Type_: types.Ptr{Elem: info.Type_},
		}
	}
	return types.Unknown{}
}

func getBasicLitConstType(kind token.Token) types.Info {
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
func (c *Scope) tryResolvePkgRef(sel *ast.SelectorExpr) types.Info {
	idt, ok := sel.X.(*ast.Ident)
	if !ok {
		return types.Unknown{}
	}
	imp, ok := c.File.Imports[idt.Name]
	if !ok {
		return types.Unknown{}
	}
	return types.Variable{
		PkgPath: imp,
		Name:    sel.Sel.Name,
	}
}

// (*ipc.Reader).Next or ipc.Reader.Next
func (c *Scope) tryResolvePkgType(expr ast.Expr) types.Info {
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

func (c *Scope) tryResolveVarDotField(expr ast.Expr) types.Info {
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
func (c *Scope) tryGetStructFieldType(pkgPath string, declName string, refFieldName string) types.Info {
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
	pkgScope := c.newFileScope(pkg, decl.File)
	return pkgScope.resolveInfo(foundField.Type)
}

func (c *Scope) newFileScope(pkg *edit.Package, file *edit.File) *Scope {
	return newFileScope(c.Global, pkg, file)
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
func (c *Scope) resolveDefType(def *Define) types.Info {
	if def.Index == -1 {
		// var := pkg.Name{}
		typeInfo := c.resolveInfo(def.Expr)
		if !types.IsUnknown(typeInfo) {
			return typeInfo
		}
	}
	// not supported yet
	return types.Unknown{}
}
