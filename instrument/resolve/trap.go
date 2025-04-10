package resolve

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// a or &a
func (c *Scope) trapIdent(addr *ast.UnaryExpr, idt *ast.Ident) bool {
	if isBlankName(idt.Name) || c.Has(idt.Name) {
		return false
	}
	decl := c.Global.Package.Decls[idt.Name]
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
	pkgPath := c.Global.Imports[xName]
	if pkgPath == "" {
		// X.Y where X is a variable
		decl := c.Global.PkgScopeNames[xName]
		if decl != nil {
			selType := c.resolveType(sel)
			if !types.IsUnknown(selType) {
				// if addr is nil, we cannot explicitly get sel.X's type
				// unless sel.X is a pointer
				xInfo := c.Global.ObjectInfo[sel.X]
				if variable, ok := xInfo.(types.Variable); ok {
					_, isPtr := variable.Type.(types.Ptr)
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
	pkgData := c.Global.Packages.PackageByPath[pkgPath]
	if pkgData == nil {
		return false
	}
	decl := pkgData.Decls[name]
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
	if decl.Type == nil {
		if decl.Value == nil {
			return false
		}
		typ := decl.ResolvedValue
		if typ == nil {
			typ = c.resolveType(decl.Value)
			decl.ResolvedValue = typ
			if types.IsUnknown(typ) {
				return false
			}
		}

		// must be basic type
		if _, ok := typ.(types.Basic); !ok {
			return false
		}
	}
	return true
}

func (ctx *Scope) applyVarRewrite(decl *edit.Decl, addr *ast.UnaryExpr, suffix string, startPos token.Pos, endPos token.Pos) {
	// change pkg.VarName to pkg.VarName_xgo_get()
	decl.HasCallRewrite = true

	// TODO: support address
	fileEdit := ctx.Global.File.Edit
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
