package instrument_var

import (
	"go/ast"
	"strconv"
	"strings"
)

// name -> path
type Imports map[string]string

func getFileImports(file *ast.File) Imports {
	imports := make(Imports)
	for _, impDecl := range file.Imports {
		pkgPath, err := strconv.Unquote(impDecl.Path.Value)
		if err != nil {
			continue
		}
		var localName string
		if impDecl.Name != nil {
			localName = impDecl.Name.Name
		} else {
			idx := strings.LastIndex(pkgPath, "/")
			if idx < 0 {
				localName = pkgPath
			} else {
				localName = pkgPath[idx+1:]
			}
		}
		if localName == "" || localName == "." || localName == "_" {
			continue
		}
		imports[localName] = pkgPath
	}
	return imports
}
