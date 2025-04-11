package resolve

import "go/ast"

func debugpoint() {}

func onTraverseFuncDecl(fnDecl *ast.FuncDecl) {
	if fnDecl.Name != nil && fnDecl.Name.Name == "TestPatchGenericFunc" {
		debugpoint()
	}
}
