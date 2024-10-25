//go:build !go1.18
// +build !go1.18

package code

import "go/ast"

// support generic
func (c *formatter) handleTypeParams(n *ast.FuncType) {
	// go1.17 and under do nothing
}

func (c *formatter) handleTypeParamsForTypeSpec(n *ast.TypeSpec) {
}

// support generic
func (c *formatter) handleFallback(n ast.Node) bool {
	// go1.17 and under do nothing
	return false
}
