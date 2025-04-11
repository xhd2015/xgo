package resolve

import "go/ast"

func (c *Scope) traverseFuncLit(node *ast.FuncLit) {
	if node == nil {
		return
	}
	c.traverseFunc(nil, node.Type, node.Body)
}

func (c *Scope) traverseFunc(recv *ast.FieldList, funcType *ast.FuncType, body *ast.BlockStmt) {
	// each function is a new context
	names := make(map[string]bool)
	argNames := getFuncDeclNamesNoBlank(recv, funcType)
	for _, argName := range argNames {
		names[argName] = true
	}
	scope := c.newScope()
	scope.Names = names
	scope.traverseBlockStmt(body)
}

func (c *Scope) traverseBlockStmt(node *ast.BlockStmt) {
	if node == nil {
		return
	}
	scope := c.newScope()
	n := len(node.List)
	for i := 0; i < n; i++ {
		scope.traverseStmt(node.List[i])
	}
}
