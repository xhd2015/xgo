package config_debug

import (
	"go/ast"

	"github.com/xhd2015/xgo/instrument/edit"
)

func Debugpoint() {}

// to debug, add `--debug-xgo` and `-tags dev`:
//
//	go run ./script/run-test --include go1.24.1 -tags dev -run TestFuncTab -v --debug-xgo ./runtime/test/functab

func OnTraverseFile(pkg *edit.Package, file *edit.File) {
	if file.File.Name == "x509.go" {
		Debugpoint()
	}
}

func OnCollectFileDecl(pkg *edit.Package, file *edit.File) {
	if file.File.Name == "type_alias_go_1.20_test.go" {
		Debugpoint()
	}
}

func OnTraverseFuncDecl(pkg *edit.Package, file *edit.File, fnDecl *ast.FuncDecl) {
	var funcName string
	if fnDecl.Name != nil {
		funcName = fnDecl.Name.Name
	}
	if pkg.LoadPackage.GoPackage.ImportPath == "crypto/x509" {
		if file.File.Name == "x509.go" {
			if recvNoName(fnDecl.Recv) && fnDecl.Name != nil && fnDecl.Name.Name == "Error" {
				Debugpoint()
			}
		}
	}
	if funcName == "TestTypeAliasGenericNonPtrDebug" {
		Debugpoint()
	}
}

func OnTrapFunc(pkgPath string, fnDecl *ast.FuncDecl, identityName string) {
	if identityName == "(*ConstraintViolationError).Error" {
		Debugpoint()
	}
}

func recvNoName(recv *ast.FieldList) bool {
	return recv != nil && len(recv.List) == 1 && len(recv.List[0].Names) == 0
}

func DebugExprStr(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return DebugExprStr(expr.X) + "." + expr.Sel.Name
	case *ast.ParenExpr:
		return "(" + DebugExprStr(expr.X) + ")"
	}
	return ""
}

func AfterSelectorResolve(expr ast.Expr) {
	if DebugExprStr(expr) == "reader2222.Reader" {
		Debugpoint()
	}
}

func OnRewriteVarDefAndRefs(pkgPath string, file *edit.File, decl *edit.Decl) {
	fileName := file.File.Name
	var declName string
	if decl.Ident != nil {
		declName = decl.Ident.Name
	}

	if fileName == "type_ref_multiple_times_test.go" {
		if declName == "testMap" {
			Debugpoint()
		}
	}
	if fileName == "tree.go" {
		if declName == "Tree" {
			Debugpoint()
		}
	}
}
