package patch

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

// NOTE: new func is always inserted after a position
// because a function or variable may have some
// special directive like go:embed, go:linkname...
// we cannot add anything before them.
func GetFuncInsertPosition(file *ast.File) token.Pos {
	// find first non-import declaration
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			return decl.End()
		}
		if genDecl.Tok != token.IMPORT {
			return genDecl.End()
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
	return fileEnd
}

func AddImport(edit *goedit.Edit, file *ast.File, name string, pkgPath string) {
	edit.Insert(file.Name.End(), fmt.Sprintf(";import %s %q", name, pkgPath))
}

func AddCode(edit *goedit.Edit, pos token.Pos, code string) {
	edit.Insert(pos, ";"+code)
}
