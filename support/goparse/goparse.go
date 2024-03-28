package goparse

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func Parse(file string) (code []byte, ast *ast.File, fset *token.FileSet, err error) {
	code, err = os.ReadFile(file)
	if err != nil {
		return
	}
	ast, fset, err = ParseFileCode(file, code)
	return
}

// file is optional
func ParseFileCode(file string, code []byte) (ast *ast.File, fset *token.FileSet, err error) {
	fset = token.NewFileSet()
	ast, err = parser.ParseFile(fset, file, code, parser.ParseComments)
	return
}

func AddMissingPackage(code string, pkgName string) string {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			return code
		}
	}
	return "package " + pkgName + ";" + code
}
