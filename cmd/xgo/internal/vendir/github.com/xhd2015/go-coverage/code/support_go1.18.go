//go:build go1.18
// +build go1.18

package code

import "go/ast"

// actually go1.17 has generic support, through +build typeparams,see:
//     https://github.com/golang/go/blob/go1.17.13/src/go/ast/ast_typeparams.go
//     https://github.com/golang/go/blob/go1.17.13/src/go/ast/ast_notypeparams.go

// support generic
func (c *formatter) handleTypeParams(n *ast.FuncType) {
	// go1.17 and under do nothing
	if n.TypeParams == nil {
		return
	}
	c.add("[")
	c.cleanCode(n.TypeParams)
	c.add("]")
}

// example:
//  type List[T any] struct {
//  	next  *List[T]
//  	value T
//  }
func (c *formatter) handleTypeParamsForTypeSpec(n *ast.TypeSpec) {
	// go1.17 and under do nothing
	if n.TypeParams == nil {
		return
	}
	c.add("[")
	c.cleanCode(n.TypeParams)
	c.add("]")
}

// support generic
func (c *formatter) handleFallback(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.IndexListExpr:
		// added in go1.18, example: reverse[int,int64]([]int{1, 2, 3, 4, 5})
		c.cleanCode(n.X) // the func name
		if n.Lbrack.IsValid() {
			c.add("[")
		}
		c.cleanExprs(n.Indices, ",")
		if n.Rbrack.IsValid() {
			c.add("]")
		}
		return true
	}
	return false
}
