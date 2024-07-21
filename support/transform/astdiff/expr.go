package astdiff

import (
	"fmt"
	"go/ast"
	"go/token"
)

func ExprSame(a ast.Expr, b ast.Expr) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	switch a := a.(type) {
	case *ast.Ident:
		b, ok := b.(*ast.Ident)
		if !ok {
			return false
		}
		return a.Name == b.Name
	case *ast.BasicLit:
		b, ok := b.(*ast.BasicLit)
		if !ok {
			return false
		}
		return basicLitSame(a, b)
	case *ast.SelectorExpr:
		b, ok := b.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		return identSame(a.Sel, b.Sel) && ExprSame(a.X, b.X)
	case *ast.CallExpr:
		b, ok := b.(*ast.CallExpr)
		if !ok {
			return false
		}
		if !ExprSame(a.Fun, b.Fun) {
			return false
		}
		if !ExprsSame(a.Args, b.Args) {
			return false
		}
		if (a.Ellipsis == token.NoPos) != (b.Ellipsis == token.NoPos) {
			return false
		}
		return true
	case *ast.BinaryExpr:
		b, ok := b.(*ast.BinaryExpr)
		if !ok {
			return false
		}
		if a.Op != b.Op {
			return false
		}
		return ExprSame(a.X, b.X) && ExprSame(a.Y, b.Y)
	case *ast.StarExpr:
		b, ok := b.(*ast.StarExpr)
		if !ok {
			return false
		}
		return ExprSame(a.X, b.X)
	case *ast.FuncLit:
		b, ok := b.(*ast.FuncLit)
		if !ok {
			return false
		}
		return funcTypeSame(a.Type, b.Type) && blockStmtSame(a.Body, b.Body)
	case *ast.FuncType:
		b, ok := b.(*ast.FuncType)
		if !ok {
			return false
		}
		return funcTypeSame(a, b)
	default:
		panic(fmt.Errorf("unrecognized: %T", a))
	}
	return false
}

func ExprsSame(a []ast.Expr, b []ast.Expr) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if !ExprSame(e, b[i]) {
			return false
		}
	}
	return true
}

func FuncDeclSame(a *ast.FuncDecl, b *ast.FuncDecl) bool {
	return funcDeclSame(a, b, false)
}

func FuncDeclSameIgnoreBody(a *ast.FuncDecl, b *ast.FuncDecl) bool {
	return funcDeclSame(a, b, true)
}

func funcDeclSame(a *ast.FuncDecl, b *ast.FuncDecl, ignoreBody bool) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	if !fieldListSame(a.Recv, b.Recv) {
		return false
	}
	if !identSame(a.Name, b.Name) {
		return false
	}
	if !funcTypeSame(a.Type, b.Type) {
		return false
	}
	if !ignoreBody && !blockStmtSame(a.Body, b.Body) {
		return false
	}
	return true
}
