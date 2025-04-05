package instrument_var

import "go/ast"

func (c *BlockContext) traverseFuncLit(node *ast.FuncLit, pkgScopeNames PkgScopeNames, imports Imports) {
	if node == nil {
		return
	}
	c.traverseFunc(nil, node.Type, node.Body, pkgScopeNames, imports)
}

func (c *BlockContext) traverseFunc(recv *ast.FieldList, funcType *ast.FuncType, body *ast.BlockStmt, pkgScopeNames PkgScopeNames, imports Imports) {
	// each function is a new context
	ctx := &BlockContext{
		Global: c.Global,
		Names:  make(map[string]bool),
	}
	argNames := getFuncDeclNamesNoBlank(recv, funcType)
	for _, argName := range argNames {
		ctx.Names[argName] = true
	}
	ctx.traverseBlockStmt(body, pkgScopeNames, imports)
}

// imports: map[string]string, key is local name, value is import path
func (c *BlockContext) traverseBlockStmt(node *ast.BlockStmt, pkgScopeNames PkgScopeNames, imports Imports) {
	if node == nil {
		return
	}
	subCtx := &BlockContext{
		Parent: c,
		Global: c.Global,
	}
	n := len(node.List)
	for i := 0; i < n; i++ {
		subCtx.traverseStmt(node.List[i], pkgScopeNames, imports)
	}
}
