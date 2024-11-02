package ast

import (
	"go/ast"
)

type file struct {
	relPath   string
	ast       *ast.File
	content   []byte
	syntaxErr error
}

var _ File = (*file)(nil)

func (c *file) RelPath() string {
	return c.relPath
}

func (c *file) Ast() *ast.File {
	return c.ast
}

func (c *file) Content() []byte {
	return c.content
}

func (c *file) SyntaxError() error {
	return c.syntaxErr
}
