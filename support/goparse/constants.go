package goparse

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/support/strutil"
)

// ExtractConstants extracts constants from a go file
// either fileName or content is required
func ExtractConstants(fileName string, content string) (map[string]string, error) {
	if fileName == "" && content == "" {
		return nil, fmt.Errorf("either fileName or content is required")
	}
	file, _, err := ParseFileCode(fileName, strutil.ToReadonlyBytes(content))
	if err != nil {
		return nil, err
	}

	constants := make(map[string]string)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			constSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			var names []string
			for _, name := range constSpec.Names {
				names = append(names, name.Name)
			}
			var values []string
			for _, value := range constSpec.Values {
				var valueLit string
				if lit, ok := value.(*ast.BasicLit); ok {
					valueLit = lit.Value
				} else {
					return nil, fmt.Errorf("unknown const spec value: %T %v", value, value)
				}
				values = append(values, valueLit)
			}
			if len(names) != len(values) {
				return nil, fmt.Errorf("mismatch decl: names=%v,values=%v", names, values)
			}

			for i, name := range names {
				constants[name] = values[i]
			}
		}
	}
	return constants, nil
}
