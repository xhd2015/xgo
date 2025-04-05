package instrument_var

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/edit"
)

func (ctx *BlockContext) trapIdent(addr *ast.UnaryExpr, idt *ast.Ident, pkgScopeNames PkgScopeNames, imports Imports) bool {
	if isBlankName(idt.Name) || ctx.Has(idt.Name) {
		return false
	}
	decl := ctx.Global.Package.Decls[idt.Name]
	if decl == nil {
		return false
	}
	if decl.Kind != edit.DeclKindVar || decl.Type == nil {
		return false
	}

	ctx.applyVarRewrite(decl, addr, idt.Pos(), idt.End())
	return true
}

// if is &pkg.VarName, return &pkg.VarName_xgo_get_addr()
func (ctx *BlockContext) trapSelector(addr *ast.UnaryExpr, sel *ast.SelectorExpr, pkgScopeNames PkgScopeNames, imports Imports) bool {
	// form: pkg.var
	xNode, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	xName := xNode.Name
	if isBlankName(xName) || ctx.Has(xName) {
		// local name
		return false
	}
	// import path
	pkgPath := imports[xName]
	if pkgPath == "" {
		// sel.X = ctx.trapValueNode(xNode, globaleNames)
		return false
	}
	// var explicitType syntax.Expr

	name := sel.Sel.Name

	// pkgPath like "fmt", other libs are ignored
	pkgData := ctx.Global.Packages.PackageByPath[pkgPath]
	if pkgData == nil {
		return false
	}
	decl := pkgData.Decls[name]
	if decl == nil {
		return false
	}
	if decl.Kind != edit.DeclKindVar || decl.Type == nil {
		return false
	}
	ctx.applyVarRewrite(decl, addr, sel.Pos(), sel.Sel.End())
	return true
}

func (ctx *BlockContext) applyVarRewrite(decl *edit.Decl, addr *ast.UnaryExpr, startPos token.Pos, endPos token.Pos) {
	// change pkg.VarName to pkg.VarName_xgo_get()
	decl.HasCallRewrite = true

	// TODO: support address
	fileEdit := ctx.Global.File.Edit
	if addr == nil {
		fileEdit.Insert(endPos, "_xgo_get()")
	} else {
		// delete &
		fileEdit.Delete(addr.OpPos, startPos)
		fileEdit.Insert(endPos, "_xgo_get_addr()")
	}
}
