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
	return c.trapExpr(addr, idt)
}

func debugpoint() {}

// if is &pkg.VarName, return &pkg.VarName_xgo_get_addr()
func (c *Scope) trapSelector(addr *ast.UnaryExpr, sel *ast.SelectorExpr) bool {
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
	return c.trapExpr(addr, expr)
}

func (c *Scope) trapExpr(addr *ast.UnaryExpr, expr ast.Expr) bool {
	expr = deparen(expr)
	exprInfo := c.resolveInfo(expr)
	obj, ok := exprInfo.(types.Object)
	if !ok {
		// xInfo could be ImportPath
		impPath, ok := exprInfo.(types.ImportPath)
		if !ok {
			return false
		}
		_ = impPath
		return false
	}

	pkgVar, ok := obj.(types.PkgVariable)
	if !ok {
		return false
	}

	decl := c.getPkgNameDecl(pkgVar.PkgPath, pkgVar.Name)
	if decl == nil {
		return false
	}

	if !c.canDeclareType(decl) {
		return false
	}

	var isPtr bool
	if addr != nil {
		isPtr = true
	} else {
		isPtr = types.IsPointer(pkgVar.Type())
	}
	c.applyVarRewrite(decl, addr, getSuffix(isPtr), expr.End())
	return true
}

func (ctx *Scope) applyVarRewrite(decl *edit.Decl, addr *ast.UnaryExpr, suffix string, nameEndPos token.Pos) {
	// change pkg.VarName to pkg.VarName_xgo_get()
	decl.HasCallRewrite = true

	fileEdit := ctx.File.File.Edit
	if addr != nil {
		// delete &
		fileEdit.Delete(addr.Pos(), addr.X.Pos())
	}
	fileEdit.Insert(nameEndPos, suffix)
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

func getSuffix(isPtr bool) string {
	if isPtr {
		return "_xgo_get_addr()"
	}
	return "_xgo_get()"
}

type UseContext struct {
	fset         *token.FileSet
	pkgPath      string
	file         *edit.File
	pos          token.Pos
	deferedEdits []func()
}

func (c *UseContext) useTypeInFile(typ types.Type) (res string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "useTypeInFile: %v\n", r)
			res = ""
		}
	}()
	typeCode := c.doUseTypeInFile(typ)
	// apply edit if no panic
	for _, edit := range c.deferedEdits {
		edit()
	}
	return typeCode
}

func (c *UseContext) doUseTypeInFile(typ types.Type) (res string) {
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
		c.deferedEdits = append(c.deferedEdits, func() {
			patch.AddImport(c.file.Edit, c.file.File.Syntax, pkgRef, typ.PkgPath)
		})
		return fmt.Sprintf("%s.%s", pkgRef, typ.Name)
	case types.Ptr:
		return "*" + c.useTypeInFile(typ.Elem)
	case types.Signature:
		panic("TODO signature")
	default:
		panic(fmt.Sprintf("unsupported type: %T", typ))
	}
}
