package resolve

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"

	"github.com/xhd2015/xgo/instrument/config"
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

// a -> look for definition
// a.b -> look for a's definition first, a could a package name, a type or a variable.
// a.b.c -> look for a.b's definition
func (c *Scope) resolveInfo(expr ast.Expr) types.Info {
	if expr == nil {
		return types.Unknown{}
	}
	info, ok := c.Global.ExprInfo[expr]
	if ok {
		return info
	}
	lazy := types.Lazy(func() types.Info {
		return info
	})
	c.setInfo(expr, lazy)
	info = c.doResolveInfo(expr)
	c.setInfo(expr, info)
	return info
}

// var debugLevel = 0

func (c *Scope) resolveType(expr ast.Expr) types.Type {
	// debugLevel++
	// defer func() {
	// 	debugLevel--
	// }()
	// if debugLevel > 10 {
	// 	fmt.Fprintf(os.Stderr, "resolveType: %T\n", expr)
	// }
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
						valInfo := c.resolveInfo(decl.Value)
						if obj, ok := valInfo.(types.Object); ok {
							varType = obj.Type()
						}
					}
					if types.IsUnknown(varType) {
						return types.Unknown{}
					}
					return types.PkgVariable{
						PkgPath: c.Package.PkgPath(),
						Name:    name,
						Type_:   varType,
					}
				} else if decl.Kind == edit.DeclKindFunc {
					var signature types.Signature
					if decl.FuncDecl != nil && decl.FuncDecl.Syntax != nil && decl.FuncDecl.Syntax.Type != nil {
						typ := c.resolveType(decl.FuncDecl.Syntax.Type)
						if typ, ok := typ.(types.Signature); ok {
							signature = typ
						}
					}
					return types.PkgFunc{
						PkgPath:   c.Package.PkgPath(),
						Name:      name,
						Signature: signature,
					}
				}
				return types.Unknown{}
			}
			pkgPath, ok := c.File.Imports[expr.Name]
			if ok {
				// this only refers to a package
				return types.ImportPath(pkgPath)
			}
			if expr.Name == "nil" {
				return types.UntypedNil{}
			}
			switch expr.Name {
			case "new", "make", "cap", "len", "copy", "append", "delete", "close", "complex", "real", "imag", "panic", "recover":
				return types.OperationName(expr.Name)
			}
			return parseBasicName(expr.Name)
		}
		return types.Unknown{}
	case *ast.SelectorExpr:
		subType := c.resolveSelectorSubType(expr)
		c.setInfo(expr.Sel, subType)
		return subType
	case *ast.StarExpr:
		xInfo := c.resolveInfo(expr.X)
		if types.IsUnknown(xInfo) {
			return types.Unknown{}
		}
		if typ, ok := xInfo.(types.Type); ok {
			return types.PtrType{Elem: typ}
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
		info := c.resolveInfo(expr.X)
		if types.IsUnknown(info) {
			return types.Unknown{}
		}
		switch info := info.(type) {
		case types.Object:
			if expr.Op == token.AND {
				return types.Pointer{
					Value: info,
				}
			}
		}
		return types.Unknown{}
	case *ast.CompositeLit:
		// generic
		litType := expr.Type
		if idxExpr, ok := expr.Type.(*ast.IndexExpr); ok {
			litType = idxExpr.X
		}

		typ := c.resolveType(litType)
		if types.IsUnknown(typ) {
			return types.Unknown{}
		}
		return types.Literal{
			Type_: typ,
		}
	case *ast.IndexExpr:
		info := c.resolveInfo(expr.X)
		// if info is a Type, then this is a
		// generic type
		switch info := info.(type) {
		case types.Type:
			return info
		case types.PkgFunc:
			return info
		case types.PkgTypeMethod:
			return info
		}

		// otherwise this yields an element type
		// from a slice or array
		// TODO: support array
		return types.Unknown{}
	case *ast.BasicLit:
		typ := getBasicLitConstType(expr.Kind)
		if types.IsUnknown(typ) {
			return types.Unknown{}
		}
		return types.Literal{
			Type_: typ,
		}
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
	case *ast.ParenExpr:
		// x := exec.Command("")
		// args := (x).Args  --> x is a pointer
		return c.resolveInfo(expr.X)
	case *ast.MapType:
		return types.Map{}
	case *ast.ArrayType:
		return types.Array{
			Elem: c.resolveType(expr.Elt),
		}
	case *ast.InterfaceType:
		return types.Interface{}
	case *ast.ChanType:
		return types.Chan{}
	case *ast.FuncType:
		// we don't need params
		// params := make([]types.Type, len(expr.Params.List))
		// for i, param := range expr.Params.List {
		// 	params[i] = c.resolveType(param.Type)
		// }
		var results []types.Type
		if expr.Results != nil {
			results = make([]types.Type, len(expr.Results.List))
			for i, result := range expr.Results.List {
				results[i] = c.resolveType(result.Type)
			}
		}
		return types.Signature{
			Results: results,
		}
	case *ast.FuncLit:
		typ := c.resolveType(expr.Type)
		if types.IsUnknown(typ) {
			return types.Unknown{}
		}
		return types.Literal{
			Type_: typ,
		}
	case *ast.CallExpr:
		// TODO more cases
		fnInfo := c.resolveInfo(expr.Fun)
		if op, ok := fnInfo.(types.OperationName); ok {
			if op == "new" {
				if len(expr.Args) != 1 {
					return types.Unknown{}
				}
				argType := c.resolveType(expr.Args[0])
				if types.IsUnknown(argType) {
					return types.Unknown{}
				}
				return types.Pointer{
					Value: types.Value{
						Type_: argType,
					},
				}
			}
			// TODO more
			return types.Unknown{}
		}
		if typ, ok := fnInfo.(types.Type); ok {
			// if len(expr.Args) != 1 {
			// 	return types.Unknown{}
			// }
			// we don't need the value information
			// this is just a type conversion
			return types.Value{
				Type_: typ,
			}
		}
		if obj, ok := fnInfo.(types.Object); ok {
			// must be function call
			signature, ok := obj.Type().Underlying().(types.Signature)
			if !ok {
				return types.Unknown{}
			}
			// tuple
			tuple := make(types.TupleValue, len(signature.Results))
			for i := range signature.Results {
				tuple[i] = types.Value{
					Type_: signature.Results[i],
				}
			}
			if len(tuple) == 1 {
				return tuple[0]
			}
			return tuple
		}
		return types.Unknown{}
	case *ast.TypeAssertExpr:
		// TODO
		return types.Unknown{}
	default:
		info, ok := c.resolveIndexListExpr(expr)
		if ok {
			return info
		}
	}
	if config.DEBUG {
		fmt.Fprintf(os.Stderr, "unresolved expr: %T\n", expr)
	}
	return types.Unknown{}
}

func parseBasicName(name string) types.Type {
	switch name {
	case "string", "byte":
		return types.Basic(name)
	case "int", "int8", "int16", "int32", "int64":
		return types.Basic(name)
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return types.Basic(name)
	case "float32", "float64":
		return types.Basic(name)
	case "bool":
		return types.Basic(name)
	case "complex64", "complex128":
		return types.Basic(name)
		// TODO: "error"
	}
	return types.Unknown{}
}

func forAllParenNestedExpr(expr ast.Expr, fn func(expr ast.Expr)) {
	fn(expr)
	switch expr := expr.(type) {
	case *ast.ParenExpr:
		forAllParenNestedExpr(expr.X, fn)
	}
}

func (c *Scope) setInfo(expr ast.Expr, info types.Info) {
	if c.Global.ExprInfo == nil {
		c.Global.ExprInfo = make(map[ast.Expr]types.Info, 1)
	}
	c.Global.ExprInfo[expr] = info
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
	decl, ok := c.Global.NamedTypeToDecl[PkgName{
		PkgPath: pkgPath,
		Name:    name,
	}]
	if ok {
		return decl
	}
	pkg := c.Global.Packages.GetPackage(pkgPath)
	if pkg == nil {
		return nil
	}
	return pkg.Decls[name]
}

func (c *Scope) loadPackage(pkgPath string) *edit.Package {
	pkg, _, err := c.Global.Packages.LoadPackage(pkgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %s\n", pkgPath, err)
		return nil
	}
	if pkg == nil {
		return nil
	}
	CollectDecls(pkg)
	return pkg
}

func deref(typ types.Info) types.Type {
	if ptr, ok := typ.(types.PtrType); ok {
		return ptr.Elem
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
		scope := c
		if def.Split {
			scope = c.Parent
		}
		typeInfo := scope.resolveInfo(def.Expr)
		if !types.IsUnknown(typeInfo) {
			return typeInfo
		}
	}
	// not supported yet
	return types.Unknown{}
}

func (c *Scope) resolveSelectorSubType(expr *ast.SelectorExpr) types.Info {
	selName := expr.Sel.Name
	xInfo := c.resolveInfo(expr.X)
	if types.IsUnknown(xInfo) {
		return types.Unknown{}
	}
	switch xInfo := xInfo.(type) {
	case types.ImportPath:
		// pkg.Name
		pkgPath := string(xInfo)
		pkg := c.loadPackage(pkgPath)
		if pkg == nil {
			return types.Unknown{}
		}
		if decl := pkg.Decls[selName]; decl != nil {
			pkgScope := c.newFileScope(pkg, decl.File)
			return pkgScope.resolveInfo(decl.Ident)
		}
	case types.Object:
		// var.Field
		// if xType resolves to a Decl,
		// check if Name resolves to method of that Decl
		valType := xInfo.Type()
		if valType == nil || types.IsUnknown(valType) {
			return types.Unknown{}
		}
		return c.resolveTypeField(valType, selName, func() {
			// update the nested expr type to be pointer
			forAllParenNestedExpr(expr.X, func(expr ast.Expr) {
				info := c.resolveInfo(expr)
				if obj, ok := info.(types.Object); ok {
					c.Global.ExprInfo[expr] = types.Pointer{
						Value: obj,
					}
				}

			})
		})
	case types.Type:
		// (*ipc.Reader).Next or ipc.Reader.Next
		// resolves to Value
		// Type.Something resolves to Method or Value
		return c.resolveTypeField(xInfo, selName, nil)
	}
	return types.Unknown{}
}

func (c *Scope) resolveTypeField(valType types.Type, fieldName string, onDetectedPointer func()) types.Info {
	var isPtr bool
	resolveType := valType
	if ptrType, ok := valType.(types.PtrType); ok {
		// *T
		isPtr = true
		resolveType = ptrType.Elem
	}
	// in go, a named type cannot be pointer
	if namedType, ok := resolveType.(types.NamedType); ok {
		decl := c.getPkgNameDecl(namedType.PkgPath, namedType.Name)
		var method *edit.FuncDecl
		if decl != nil {
			method = decl.Methods[fieldName]
		}
		if method != nil {
			if !isPtr && method.RecvPtr {
				// update expr.X's type to be ptr
				// TODO: if is paren, deparen all
				if onDetectedPointer != nil {
					onDetectedPointer()
				}
			}
			return types.PkgTypeMethod{
				Name: fieldName,
				Recv: namedType,
				// TODO params
			}
		} else {
			// check fields
			ud, ok := valType.Underlying().(types.Struct)
			if ok {
				var fieldType types.Type
				for _, field := range ud.Fields {
					if field.Name == fieldName {
						fieldType = field.Type
						break
					}
				}
				if fieldType != nil && !types.IsUnknown(fieldType) {
					return fieldType
				}
			}
		}
	}
	return types.Unknown{}
}
