package resolve

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"

	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// a or &a
func (c *Scope) trapIdent(addr *ast.UnaryExpr, idt *ast.Ident) bool {
	if isBlankName(idt.Name) || c.Has(idt.Name) {
		return false
	}
	decl := c.Package.Decls[idt.Name]
	if decl == nil {
		return false
	}
	if decl.Kind != edit.DeclKindVar {
		return false
	}
	if !c.canDeclareType(decl) {
		return false
	}

	c.applyVarRewrite(decl, addr, getSuffix(addr != nil), idt.Pos(), idt.End())
	return true
}

// if is &pkg.VarName, return &pkg.VarName_xgo_get_addr()
func (c *Scope) trapSelector(addr *ast.UnaryExpr, sel *ast.SelectorExpr) bool {
	// form: pkg.var
	xNode, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	xName := xNode.Name
	if isBlankName(xName) || c.Has(xName) {
		// empty or local name
		return false
	}
	// import path
	pkgPath := c.File.Imports[xName]
	if pkgPath == "" {
		// X.Y where X is a variable
		decl := c.Package.Decls[xName]
		if decl != nil {
			if !c.canDeclareType(decl) {
				return false
			}
			selType := c.resolveInfo(sel)
			if !types.IsUnknown(selType) {
				// if addr is nil, we cannot explicitly get sel.X's type
				// unless sel.X is a pointer
				xInfo := c.Global.ExprInfo[sel.X]
				if variable, ok := xInfo.(types.Variable); ok {
					_, isPtr := variable.Type().(types.Ptr)
					c.applyVarRewrite(decl, nil, getSuffix(isPtr), sel.X.End(), sel.X.End())
					return true
				}

			}
		}

		// sel.X = ctx.trapValueNode(xNode, globaleNames)
		return false
	}
	// var explicitType syntax.Expr

	// resolve sel.X's type, which is based on whether
	// sel.X is a pointer
	name := sel.Sel.Name

	// pkgPath like "fmt", other libs are ignored
	pkg := c.Global.Packages.GetPackage(pkgPath)
	if pkg == nil {
		return false
	}
	decl := pkg.Decls[name]
	if decl == nil {
		return false
	}
	if decl.Kind != edit.DeclKindVar {
		return false
	}
	if !c.canDeclareType(decl) {
		return false
	}

	suffix := getSuffix(addr != nil)
	c.applyVarRewrite(decl, addr, suffix, sel.Pos(), sel.Sel.End())
	return true
}

func (c *Scope) canDeclareType(decl *edit.Decl) bool {
	if decl.Type != nil {
		return true
	}
	if decl.Value == nil {
		return false
	}
	resolvedValue := decl.ResolvedValue
	if resolvedValue != nil {
		return decl.ResolvedValueTypeCode != ""
	}
	info := c.resolveInfo(decl.Value)
	if types.IsUnknown(info) {
		return false
	}
	obj, ok := info.(types.Object)
	if !ok {
		return false
	}
	objType := obj.Type()
	if types.IsUnknown(objType) {
		return false
	}
	useContext := &UseContext{
		fset:    c.Global.Packages.Fset(),
		pkgPath: c.Package.PkgPath(),
		file:    c.File.File,
		pos:     decl.Ident.Pos(),
	}
	typeCode := useContext.useTypeInFile(objType)
	// must be basic type
	decl.ResolvedValue = obj
	decl.ResolvedValueTypeCode = typeCode
	return typeCode != ""
}

func (ctx *Scope) applyVarRewrite(decl *edit.Decl, addr *ast.UnaryExpr, suffix string, startPos token.Pos, endPos token.Pos) {
	// change pkg.VarName to pkg.VarName_xgo_get()
	decl.HasCallRewrite = true

	fileEdit := ctx.File.File.Edit
	if addr != nil {
		// delete &
		fileEdit.Delete(addr.OpPos, startPos)
	}
	fileEdit.Insert(endPos, suffix)
}

func getSuffix(isPtr bool) string {
	if isPtr {
		return "_xgo_get_addr()"
	}
	return "_xgo_get()"
}

type UseContext struct {
	fset    *token.FileSet
	pkgPath string
	file    *edit.File
	pos     token.Pos
}

func (c *UseContext) useTypeInFile(typ types.Type) (res string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "useTypeInFile: %v\n", r)
			res = ""
		}
	}()
	switch typ := typ.(type) {
	case types.Basic:
		return string(typ)
	case types.NamedType:
		// check if pkg path is local
		if typ.PkgPath == c.pkgPath {
			return typ.Name
		}
		pos := c.fset.Position(c.pos)
		pkgRef := fmt.Sprintf("__xgo_var_ref_%d_%d", pos.Line, pos.Column)
		patch.AddImport(c.file.Edit, c.file.File.Syntax, pkgRef, typ.PkgPath)
		return fmt.Sprintf("%s.%s", pkgRef, typ.Name)
	case types.Ptr:
		return "*" + c.useTypeInFile(typ.Elem)
	case types.Signature:
		panic("TODO signature")
	default:
		panic(fmt.Sprintf("unsupported type: %T", typ))
	}
}
