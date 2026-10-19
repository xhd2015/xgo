package resolve

import (
	"go/ast"
	"go/token"
	"sync"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/config/config_debug"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// a or &a
func (c *Scope) collectIdent(addr *ast.UnaryExpr, idt *ast.Ident) bool {
	if isBlankName(idt.Name) || c.Has(idt.Name) {
		return false
	}
	// only a variable (or constant) of the package resolves to a
	// PkgVariable: skip resolving functions, types, imports and builtins
	if decl := c.Package.Decls[idt.Name]; decl == nil || (decl.Kind != edit.DeclKindVar && decl.Kind != edit.DeclKindConst) {
		return false
	}
	if config.DEBUG {
		config_debug.OnCollectVarRef(c.File.File.File.Name, idt.Name)
	}
	return c.collectVarRef(addr, idt, nil)
}

// if is &pkg.VarName, return &pkg.VarName_xgo_get_addr()
func (c *Scope) collectSelector(addr *ast.UnaryExpr, sel *ast.SelectorExpr) bool {
	if !c.mayReferPkgVar(sel) {
		return false
	}
	// parse the whole A.B to determine A's type
	selInfo := c.resolveInfo(sel)
	if types.IsUnknown(selInfo) {
		return false
	}

	// pkg.A
	// if this is a package variable, we need to trap the variable
	// instead of the selector
	expr := sel.X
	var originalSel *ast.SelectorExpr
	_, ok := selInfo.(types.PkgVariable)
	if ok {
		expr = sel.Sel
		// This is pkg.VarName, not variable.Field
		originalSel = nil
	} else {
		// This is variable.Field
		originalSel = sel
	}

	// selInfo should be an object
	return c.collectVarRef(addr, expr, originalSel)
}

func (c *Scope) collectVarRef(addr *ast.UnaryExpr, expr ast.Expr, originalSel *ast.SelectorExpr) bool {
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
	if _, allow := config.CheckInstrument(pkgPath); !allow {
		return false
	}

	decl := c.getPkgNameDecl(pkgPath, varName)
	if decl == nil {
		return false
	}
	if !c.tryResolveDeclType(pkgPath, decl) {
		return false
	}

	// Check if this is &variable.Field pattern
	fieldAccess := false
	if addr != nil && originalSel != nil {
		// This is &variable.Field pattern
		// For pointer-typed variables, we can rewrite to &variable_xgo_get().Field
		// For non-pointer variables, we cannot rewrite because variable_xgo_get() returns a value
		if !types.IsPointer(pkgVar.Type()) {
			// Cannot rewrite &nonPtrVar.Field
			return false
		}
		// For pointer variables, use _xgo_get() instead of _xgo_get_addr()
		fieldAccess = true
	}

	var needPtr bool
	if addr != nil {
		if fieldAccess {
			// For &ptrVariable.Field, we use _xgo_get() not _xgo_get_addr()
			needPtr = false
		} else {
			// force use addr
			needPtr = true
		}
	} else {
		// if original type is not pointer, check
		// if the ref is a pointer
		if !types.IsPointer(pkgVar.Type()) && isExprPtr {
			needPtr = true
		}
	}
	var nameStart token.Pos
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		nameStart = sel.Sel.Pos()
	} else {
		nameStart = expr.Pos()
	}
	c.Global.addVarRef(decl, &edit.VarRef{
		File:        c.File.File,
		Addr:        addr,
		NeedPtr:     needPtr,
		NameStart:   nameStart,
		NameEnd:     expr.End(),
		FieldAccess: fieldAccess,
	})
	return true
}

// declTypeMu guards Decl.ResolvedValueType, which the files traversed in
// parallel share.
var declTypeMu sync.Mutex

func (c *Scope) tryResolveDeclType(pkgPath string, decl *edit.Decl) bool {
	if decl.Type != nil {
		return true
	}
	if decl.Value == nil {
		return false
	}
	// func literal is perfectly fine without extra
	// effort
	if _, ok := decl.Value.(*ast.FuncLit); ok {
		return true
	}
	declTypeMu.Lock()
	resolved := decl.ResolvedValueType
	declTypeMu.Unlock()
	if resolved == nil {
		// workers racing here compute the same type, see
		// resolveDeclValueType; keep the first one
		resolved = c.resolveDeclValueType(pkgPath, decl)
		declTypeMu.Lock()
		if decl.ResolvedValueType == nil {
			decl.ResolvedValueType = resolved
		}
		resolved = decl.ResolvedValueType
		declTypeMu.Unlock()
	}
	return !types.IsUnknown(resolved)
}

// resolveDeclValueType resolves the type of the value of decl within the
// file declaring it, whose imports, package and scopes differ from those of
// the referring function in general. The result is cached on decl for every
// file referring to it, so it must not depend on the referring file: the
// value is resolved in a scope of its own, with caches of its own (those of
// the referring file can hold resolutions in progress).
func (c *Scope) resolveDeclValueType(pkgPath string, decl *edit.Decl) types.Type {
	if decl.File == nil {
		return types.Unknown{}
	}
	pkg := c.Global.Packages.GetPackage(pkgPath)
	if pkg == nil {
		return types.Unknown{}
	}
	global := &GlobalScope{
		Packages: c.Global.Packages,
		// resolving records nothing, but must not touch what the
		// referring file records either
		Recorder: &Recorder{},
		effects:  &fileEffects{},
	}
	declScope := newFileScope(global, pkg, decl.File, c.PkgDepth+1)
	obj, ok := declScope.resolveInfo(decl.Value).(types.Object)
	if !ok {
		return types.Unknown{}
	}
	if objType := obj.Type(); objType != nil && !types.IsUnknown(objType) {
		return objType
	}
	return types.Unknown{}
}

// mayReferPkgVar reports whether resolving sel can find a package variable
// to rewrite, judging from the root of its selector chain: a package var or
// const v (v.F, v.F.G...) or an instrumented import pkg (pkg.Var,
// pkg.Var.F...). Resolving a chain is also what finds that a pointer method
// is called on a field of a variable (v.F.Lock()), so a chain is resolved
// whole whenever its root may be a variable. Selectors on anything else,
// like locals or packages that are not instrumented (fmt.Sprintf), cannot
// refer to one: they are most of a program and were most of the traversal.
func (c *Scope) mayReferPkgVar(sel *ast.SelectorExpr) bool {
	root := deparen(sel.X)
	for {
		x, ok := root.(*ast.SelectorExpr)
		if !ok {
			break
		}
		root = deparen(x.X)
	}
	idt, ok := root.(*ast.Ident)
	if !ok || c.Has(idt.Name) {
		return false
	}
	if decl := c.Package.Decls[idt.Name]; decl != nil {
		return decl.Kind == edit.DeclKindVar || decl.Kind == edit.DeclKindConst
	}
	return c.isInstrumentedImport(idt)
}

// isInstrumentedImport reports whether idt, not declared in the package
// or a local scope, names an import of an instrumented package.
func (c *Scope) isInstrumentedImport(idt *ast.Ident) bool {
	pkgPath, ok := c.File.Imports[idt.Name]
	if !ok {
		return false
	}
	_, allow := config.CheckInstrument(pkgPath)
	return allow
}
