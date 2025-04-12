package resolve

import (
	"go/ast"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// a or &a
func (c *Scope) collectIdent(addr *ast.UnaryExpr, idt *ast.Ident) bool {
	if isBlankName(idt.Name) || c.Has(idt.Name) {
		return false
	}
	return c.collectVarRef(addr, idt)
}

// if is &pkg.VarName, return &pkg.VarName_xgo_get_addr()
func (c *Scope) collectSelector(addr *ast.UnaryExpr, sel *ast.SelectorExpr) bool {
	// parse the whole A.B to determine A's type
	selInfo := c.resolveInfo(sel)
	if types.IsUnknown(selInfo) {
		return false
	}

	// pkg.A
	// if this is a package variable, we need to trap the variable
	// instead of the selector
	expr := sel.X
	_, ok := selInfo.(types.PkgVariable)
	if ok {
		expr = sel.Sel
	}

	// selInfo should be an object
	return c.collectVarRef(addr, expr)
}

func (c *Scope) collectVarRef(addr *ast.UnaryExpr, expr ast.Expr) bool {
	expr = deparen(expr)
	exprInfo := c.resolveInfo(expr)
	obj, ok := exprInfo.(types.Object)
	if !ok {
		return false
	}

	var isExprPtr bool
	ptrObj, ok := obj.(types.Pointer)
	if ok {
		isExprPtr = true
		obj = ptrObj.Value
	}

	// could expr be ptr to types.PackageVariable?
	pkgVar, ok := obj.(types.PkgVariable)
	if !ok {
		return false
	}
	pkgPath := pkgVar.PkgPath
	varName := pkgVar.Name
	if !config.IsPkgAllowed(pkgPath) {
		return false
	}

	decl := c.getPkgNameDecl(pkgPath, varName)
	if decl == nil {
		return false
	}
	if !c.tryResolveDeclType(decl) {
		return false
	}

	var needPtr bool
	if addr != nil {
		// force use addr
		needPtr = true
	} else {
		// if original type is not pointer, check
		// if the ref is a pointer
		if !types.IsPointer(pkgVar.Type()) && isExprPtr {
			needPtr = true
		}
	}
	decl.VarRefs = append(decl.VarRefs, &edit.VarRef{
		File:    c.File.File,
		Addr:    addr,
		NeedPtr: needPtr,
		NameEnd: expr.End(),
	})
	return true
}

func (c *Scope) tryResolveDeclType(decl *edit.Decl) bool {
	if decl.Type != nil {
		return true
	}
	if decl.Value == nil {
		return false
	}
	if decl.ResolvedValueType != nil {
		return !types.IsUnknown(decl.ResolvedValueType)
	}
	info := c.resolveInfo(decl.Value)
	if types.IsUnknown(info) {
		decl.ResolvedValueType = types.Unknown{}
		return false
	}
	obj, ok := info.(types.Object)
	if !ok {
		decl.ResolvedValueType = types.Unknown{}
		return false
	}
	objType := obj.Type()
	if types.IsUnknown(objType) {
		decl.ResolvedValueType = types.Unknown{}
		return false
	}

	decl.ResolvedValueType = objType
	return true
}
