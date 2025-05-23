package resolve

import "go/ast"

func (ctx *Scope) traverseDecl(node ast.Decl) {
	if node == nil {
		return
	}
	switch node := node.(type) {
	case *ast.GenDecl:
		for _, spec := range node.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				if spec.Name != nil {
					ctx.Add(spec.Name.Name)
				}
			case *ast.ValueSpec:
				for _, value := range spec.Values {
					ctx.traverseExpr(value)
				}
				// name after values
				for _, name := range spec.Names {
					ctx.Add(name.Name)
				}
			}
		}
	case *ast.FuncDecl:
		if node.Name != nil {
			ctx.Add(node.Name.Name)
		}
	default:
		errorUnknown("decl", node)
	}
}
