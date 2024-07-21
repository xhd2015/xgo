package astdiff

import "go/ast"

func StmtsSame(a []ast.Stmt, b []ast.Stmt) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if !StmtSame(e, b[i]) {
			return false
		}
	}
	return true
}

func StmtSame(a ast.Stmt, b ast.Stmt) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	switch a := a.(type) {
	case *ast.ReturnStmt:
		b, ok := b.(*ast.ReturnStmt)
		if !ok {
			return false
		}
		return ExprsSame(a.Results, b.Results)
	case *ast.AssignStmt:
		b, ok := b.(*ast.AssignStmt)
		if !ok {
			return false
		}

		if a.Tok != b.Tok {
			return false
		}
		return ExprsSame(a.Lhs, b.Lhs) && ExprsSame(a.Rhs, b.Rhs)
	case *ast.ExprStmt:
		b, ok := b.(*ast.ExprStmt)
		if !ok {
			return false
		}
		return ExprSame(a.X, b.X)
	}
	return false
}
