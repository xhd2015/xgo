package astdiff

import "go/ast"

func SpecSame(a ast.Spec, b ast.Spec) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	switch a := a.(type) {
	case *ast.ImportSpec:
		b, ok := b.(*ast.ImportSpec)
		if !ok {
			return false
		}
		if !identSame(a.Name, b.Name) {
			return false
		}
		if !basicLitSame(a.Path, b.Path) {
			return false
		}
		return true
	case *ast.ValueSpec:
		b, ok := b.(*ast.ValueSpec)
		if !ok {
			return false
		}
		if !identsSame(a.Names, b.Names) {
			return false
		}
		if !ExprSame(a.Type, b.Type) {
			return false
		}
		if !ExprsSame(a.Values, b.Values) {
			return false
		}
		return true
	}
	return false
}
