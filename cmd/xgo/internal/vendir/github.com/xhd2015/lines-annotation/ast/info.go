package ast

import (
	"go/ast"
	"go/token"
)

// LoadInfo is the core abstraction of an AST load
type LoadInfo interface {
	FileSet() *token.FileSet
	RangeFiles(handler func(f File) bool)
}

type File interface {
	// relative to loading root
	RelPath() string
	Ast() *ast.File
	Content() []byte

	// has syntax error
	SyntaxError() error
}
