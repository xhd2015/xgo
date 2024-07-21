package astdiff

import (
	"go/ast"
)

// TODO: generate using template

func NodeSame(a ast.Node, b ast.Node) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	switch a := a.(type) {
	case ast.Stmt:
		b, ok := b.(ast.Stmt)
		if !ok {
			return false
		}
		return StmtSame(a, b)
	case ast.Expr:
		b, ok := b.(ast.Expr)
		if !ok {
			return false
		}
		return ExprSame(a, b)
	case ast.Spec:
		b, ok := b.(ast.Spec)
		if !ok {
			return false
		}
		return SpecSame(a, b)
	case *ast.File:
		b, ok := b.(*ast.File)
		if !ok {
			return false
		}
		return FileSame(a, b)
	}
	return false
}

func FileSame(a *ast.File, b *ast.File) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if !identSame(a.Name, b.Name) {
		return false
	}
	if !DeclsSame(a.Decls, b.Decls) {
		return false
	}
	return true
}

func DeclsSame(a []ast.Decl, b []ast.Decl) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, d := range a {
		if !DeclSame(d, b[i]) {
			return false
		}
	}
	return true
}

func DeclSame(a ast.Decl, b ast.Decl) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	switch a := a.(type) {
	case *ast.GenDecl:
		b, ok := b.(*ast.GenDecl)
		if !ok {
			return false
		}
		if a.Tok != b.Tok {
			return false
		}
		return SpecsSame(a.Specs, b.Specs)
	case *ast.FuncDecl:
		b, ok := b.(*ast.FuncDecl)
		if !ok {
			return false
		}
		return FuncDeclSame(a, b)
	case *ast.BadDecl:
		_, ok := b.(*ast.BadDecl)
		if !ok {
			return false
		}
		return true
	}
	return false
}

func SpecsSame(a []ast.Spec, b []ast.Spec) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if !SpecSame(e, b[i]) {
			return false
		}
	}
	return true
}
func fieldListSame(a *ast.FieldList, b *ast.FieldList) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	return fieldsSame(a.List, b.List)
}

func fieldsSame(a []*ast.Field, b []*ast.Field) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if !fieldSame(e, b[i]) {
			return false
		}
	}
	return true
}

func fieldSame(a *ast.Field, b *ast.Field) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if !identsSame(a.Names, b.Names) {
		return false
	}
	if !ExprSame(a.Type, b.Type) {
		return false
	}
	if !basicLitSame(a.Tag, b.Tag) {
		return false
	}
	return true
}

func blockStmtSame(a *ast.BlockStmt, b *ast.BlockStmt) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	return StmtsSame(a.List, b.List)
}

func basicLitSame(a *ast.BasicLit, b *ast.BasicLit) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	if a.Kind != b.Kind {
		return false
	}
	return a.Value == b.Value
}

func identsSame(a []*ast.Ident, b []*ast.Ident) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if !identSame(e, b[i]) {
			return false
		}
	}
	return true
}

func identSame(a *ast.Ident, b *ast.Ident) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	return a == nil || a.Name == b.Name
}
