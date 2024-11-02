package code

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

func ParseCodeString(file string, content string) (fset *token.FileSet, ast *ast.File, err error) {
	fset = token.NewFileSet()
	ast, err = parser.ParseFile(fset, file, content, parser.ParseComments)
	if err != nil {
		return
	}
	return
}

func ParseCode(file string, content []byte) (fset *token.FileSet, ast *ast.File, err error) {
	fset = token.NewFileSet()
	ast, err = parser.ParseFile(fset, file, content, parser.ParseComments)
	if err != nil {
		return
	}
	return
}

func ParseFile(f string) (fset *token.FileSet, ast *ast.File, content []byte, err error) {
	fset = token.NewFileSet()
	content, err = ioutil.ReadFile(f)
	if err != nil {
		return
	}
	ast, err = parser.ParseFile(fset, f, content, parser.ParseComments)
	if err != nil {
		return
	}
	return
}
