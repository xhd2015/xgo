package resolve

import (
	"fmt"
	"go/ast"
)

func (c *Scope) traverseFuncLit(node *ast.FuncLit) {
	if node == nil {
		return
	}
	c.traverseFunc(nil, "(closure)", node.Type, node.Body)
}

func (c *Scope) traverseFunc(recv *ast.FieldList, debugFuncName string, funcType *ast.FuncType, body *ast.BlockStmt) {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("panic while traverseFunc pkg=%v, file=%v, func=%v line=%v: %v", c.Package.PkgPath(), c.File.File.File.Name, debugFuncName, c.Global.Packages.Fset().Position(body.Pos()).Line, e))
		}
	}()
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
