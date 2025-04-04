package instrument_var

import (
	"go/ast"
	"go/token"
)

func (ctx *BlockContext) traverseExprs(nodes []ast.Expr, pkgScopeNames PkgScopeNames, imports Imports) {
	for _, node := range nodes {
		ctx.traverseExpr(node, pkgScopeNames, imports)
	}
}

func (ctx *BlockContext) traverseCallExpr(callExpr *ast.CallExpr, pkgScopeNames PkgScopeNames, imports Imports) {
	if callExpr == nil {
		return
	}
	// pkg.A.Method or A.Method
	// we don't know A's type, it could be value type or pointer type
	// we can enhance this in the future by getting type info
	var isSelector bool
	fn := callExpr.Fun
	if sel, ok := fn.(*ast.SelectorExpr); ok {
		if isIdent(sel.X) || isSelectorExpr(sel.X) {
			isSelector = true
		}
	}
	if !isSelector {
		ctx.traverseExpr(fn, pkgScopeNames, imports)
	}

	for _, arg := range callExpr.Args {
		ctx.traverseExpr(arg, pkgScopeNames, imports)
	}
}

func isIdent(expr ast.Expr) bool {
	_, ok := expr.(*ast.Ident)
	return ok
}
func isSelectorExpr(expr ast.Expr) bool {
	_, ok := expr.(*ast.SelectorExpr)
	return ok
}

func (ctx *BlockContext) traverseExpr(node ast.Expr, pkgScopeNames PkgScopeNames, imports Imports) {
	if node == nil {
		return
	}

	switch node := node.(type) {
	case *ast.Ident:
		if ctx.trapIdent(nil, node, pkgScopeNames, imports) {
			return
		}
	case *ast.SelectorExpr:
		if ctx.trapSelector(nil, node, pkgScopeNames, imports) {
			return
		}
		// var lock sync.Mutex
		// lock.Lock --> lock should not be intercepted
		// because its type cannot be determined
		// see TestMockLockShouldNotCopy
		_, ok := deparen(node.X).(*ast.Ident)
		if !ok {
			ctx.traverseExpr(node.X, pkgScopeNames, imports)
		}
	case *ast.UnaryExpr:
		// take addr
		if node.Op == token.AND {
			switch x := node.X.(type) {
			case *ast.SelectorExpr:
				if ctx.trapSelector(node, x, pkgScopeNames, imports) {
					return
				}
			case *ast.Ident:
				if ctx.trapIdent(node, x, pkgScopeNames, imports) {
					return
				}
			}
		}
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
	case *ast.CallExpr:
		ctx.traverseCallExpr(node, pkgScopeNames, imports)
	case *ast.BinaryExpr:
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
		ctx.traverseExpr(node.Y, pkgScopeNames, imports)
	case *ast.ParenExpr:
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
	case *ast.BasicLit:
		// nothing
	case *ast.FuncLit:
		ctx.traverseFuncLit(node, pkgScopeNames, imports)
	case *ast.StarExpr:
		// * something
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
	case *ast.TypeAssertExpr:
		// something.(type)
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
	case *ast.CompositeLit:
		// something{...}
		for _, elem := range node.Elts {
			ctx.traverseExpr(elem, pkgScopeNames, imports)
		}
	case *ast.IndexExpr:
		// something[...], only traverse the index
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
		ctx.traverseExpr(node.Index, pkgScopeNames, imports)
	case *ast.SliceExpr:
		// something[...:...], only traverse the index
		ctx.traverseExpr(node.X, pkgScopeNames, imports)
		ctx.traverseExpr(node.Low, pkgScopeNames, imports)
		ctx.traverseExpr(node.High, pkgScopeNames, imports)
	case *ast.ChanType:
		// chan something
	case *ast.MapType:
		// map[key]value

	case *ast.StructType:
		// struct{...}
	case *ast.ArrayType:
		// [...][...]
	case *ast.InterfaceType:
		// interface{...}
	case *ast.KeyValueExpr:
		// key: value
		ctx.traverseExpr(node.Key, pkgScopeNames, imports)
		ctx.traverseExpr(node.Value, pkgScopeNames, imports)
	default:
		errorUnknown("expr", node)
	}
}

func deparen(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}
	for {
		p, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = p.X
	}
}
