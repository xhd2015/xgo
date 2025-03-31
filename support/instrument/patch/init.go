package patch

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

type PositionType int

const (
	Before PositionType = iota
	After
)

func GetFuncInsertPosition(file *ast.File) (token.Pos, PositionType) {
	// find first non-import declaration
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			return decl.Pos(), Before
		}
		if genDecl.Tok != token.IMPORT {
			return genDecl.Pos(), Before
		}
	}

	fileEnd := file.End()
	packageEnd := file.Name.End()
	if fileEnd == packageEnd {
		// packageEnd is usually used to import instrumented packages
		// if there are other imports, it is not safe to declare
		// a new function.
		// so always insert a new line to begin the function.
		//
		// however, in reality if there is only one import,
		// there is nothing to instrument actually.
		// so we panic here
		panic("nothing to instrument with empty file")
	}
	return fileEnd, After
}

func AddImport(edit *goedit.Edit, file *ast.File, name string, pkgPath string) {
	edit.Insert(file.Name.End(), fmt.Sprintf(";import %s %q", name, pkgPath))
}

func AddCode(edit *goedit.Edit, pos token.Pos, posType PositionType, code string) {
	switch posType {
	case Before:
		edit.Insert(pos, code+";")
	case After:
		edit.Insert(pos, ";"+code)
	default:
		panic(fmt.Errorf("invalid position type: %d", posType))
	}
}
