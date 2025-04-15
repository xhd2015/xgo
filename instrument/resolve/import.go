package resolve

import (
	"go/ast"
	"strconv"
	"strings"
)

// name -> path
type Imports map[string]string

func getFileImports(registry PackageRegistry, file *ast.File) Imports {
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
			var loadName string
			pkg := registry.GetPackage(pkgPath)
			if pkg != nil {
				loadName = pkg.LoadPackage.GoPackage.Name
			}
			if loadName == "" {
				name, idx := getLastPartAsName(pkgPath)
				if idx >= 0 && isVersion(name) {
					// strip version
					// see https://github.com/xhd2015/xgo/issues/317
					name, _ = getLastPartAsName(pkgPath[:idx])
				}
				loadName = name
			}
			localName = loadName
		}
		if localName == "" || localName == "." || localName == "_" {
			continue
		}
		imports[localName] = pkgPath
	}
	return imports
}

func isVersion(s string) bool {
	if !strings.HasPrefix(s, "v") {
		return false
	}
	i, err := strconv.Atoi(s[1:])
	return err == nil && i >= 0
}

func getLastPartAsName(pkgPath string) (string, int) {
	idx := strings.LastIndex(pkgPath, "/")
	if idx < 0 {
		return pkgPath, -1
	}
	return pkgPath[idx+1:], idx
}
