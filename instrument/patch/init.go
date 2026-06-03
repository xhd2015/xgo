package patch

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

func GetFuncInsertPosition(file *ast.File) token.Pos {
	for i, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			return decl.End()
		}
		if genDecl.Tok != token.IMPORT {
			return GenDeclSafeEnd(file, i)
		}
	}

	fileEnd := file.End()
	packageEnd := file.Name.End()
	if fileEnd == packageEnd {
		panic("nothing to instrument with empty file")
	}
	return fileEnd
}

func GenDeclSafeEnd(file *ast.File, declIndex int) token.Pos {
	if declIndex+1 < len(file.Decls) && !HasGoDirective(file.Decls[declIndex+1]) {
		return file.Decls[declIndex+1].Pos()
	}
	return file.Decls[declIndex].End()
}

func HasGoDirective(decl ast.Decl) bool {
	var doc *ast.CommentGroup
	switch d := decl.(type) {
	case *ast.GenDecl:
		doc = d.Doc
	case *ast.FuncDecl:
		doc = d.Doc
	}
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.HasPrefix(c.Text, "//go:") {
			return true
		}
	}
	return false
}

func Append(edit *goedit.Edit, file *ast.File, s string) {
	edit.Buffer().Insert(len(edit.Buffer().Old()), s)
}

func AddImport(edit *goedit.Edit, file *ast.File, name string, pkgPath string) {
	edit.Insert(file.Name.End(), fmt.Sprintf(";import %s %q", name, pkgPath))
}

func AddCode(edit *goedit.Edit, pos token.Pos, code string) {
	if os.Getenv("XGO_DEBUG_USE_SEMICOLON") == "true" {
		edit.Insert(pos, ";"+code)
	} else {
		edit.Insert(pos, "\n"+code+"\n")
	}
}
