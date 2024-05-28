package test_explorer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func resolveTests(fullSubDir string) ([]*TestingItem, error) {
	files, err := os.ReadDir(fullSubDir)
	if err != nil {
		return nil, err
	}
	var results []*TestingItem
	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		if file.IsDir() {
			continue
		}
		fullFile := filepath.Join(fullSubDir, fileName)
		tests, err := parseTests(fullFile)
		if err != nil {
			return nil, err
		}
		results = append(results, tests...)
	}
	return results, nil
}

func filterItem(item *TestingItem, withCases bool) *TestingItem {
	if item == nil {
		return nil
	}

	if withCases {
		children := item.Children
		n := len(children)
		i := 0
		for j := 0; j < n; j++ {
			child := filterItem(children[j], withCases)
			if child != nil {
				children[i] = child
				i++
			}
		}
		item.Children = children[:i]
		if i == 0 && item.Kind != TestingItemKind_Case {
			return nil
		}
	} else {
		if !item.HasTestCases {
			return nil
		}
	}

	return item
}

func parseTests(file string) ([]*TestingItem, error) {
	fset, decls, err := parseTestFuncs(file)
	if err != nil {
		return nil, err
	}
	items := make([]*TestingItem, 0, len(decls))
	for _, fnDecl := range decls {
		items = append(items, &TestingItem{
			Name: fnDecl.Name.Name,
			File: file,
			Line: fset.Position(fnDecl.Pos()).Line,
			Kind: TestingItemKind_Case,
		})
	}
	return items, nil
}

func parseTestFuncs(file string) (*token.FileSet, []*ast.FuncDecl, error) {
	return parseTestFuncsCode(file, nil)
}

func parseTestFuncsCode(file string, code io.Reader) (*token.FileSet, []*ast.FuncDecl, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file, code, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	var results []*ast.FuncDecl
	for _, decl := range astFile.Decls {
		fnDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fnDecl.Name == nil {
			continue
		}
		if !strings.HasPrefix(fnDecl.Name.Name, "Test") {
			continue
		}
		if fnDecl.Body == nil {
			continue
		}
		if fnDecl.Type.Params == nil || len(fnDecl.Type.Params.List) != 1 {
			continue
		}
		results = append(results, fnDecl)
	}
	return fset, results, nil
}

func getFuncDecl(funcs []*ast.FuncDecl, name string) (*ast.FuncDecl, error) {
	for _, f := range funcs {
		if f.Name != nil && f.Name.Name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("%s not found", name)
}

func parseFuncArgs(fnDecl *ast.FuncDecl) (flags []string, args []string, err error) {
	if fnDecl.Doc == nil {
		return nil, nil, nil
	}
	for _, e := range fnDecl.Doc.List {
		if !strings.HasPrefix(e.Text, "// ") {
			continue
		}
		comment := strings.TrimPrefix(e.Text, "// ")
		xx := func(comment string, prefix string) ([]string, error) {
			if !strings.HasPrefix(comment, prefix) {
				return nil, nil
			}
			commentArgs := strings.TrimSpace(strings.TrimPrefix(comment, prefix))
			return splitArgs(commentArgs)
		}
		commentArgs, err := xx(comment, "args:")
		if err != nil {
			return nil, nil, err
		}
		args = append(args, commentArgs...)
		commentFlags, err := xx(comment, "flags:")
		if err != nil {
			return nil, nil, err
		}
		flags = append(flags, commentFlags...)
	}
	return flags, args, nil
}

func applyVars(absProjectDir string, list []string) []string {
	n := len(list)
	for i := 0; i < n; i++ {
		list[i] = strings.ReplaceAll(list[i], "$PROJECT_DIR", absProjectDir)
		list[i] = strings.ReplaceAll(list[i], "${PROJECT_DIR}", absProjectDir)
	}
	return list
}

func splitArgs(s string) ([]string, error) {
	var list []string
	runes := []rune(s)
	n := len(runes)

	var buf []rune
	for i := 0; i < n; i++ {
		ch := runes[i]
		if ch == ' ' || ch == '\t' {
			if len(buf) > 0 {
				list = append(list, string(buf))
				buf = nil
			}
			continue
		}
		buf = append(buf, ch)
	}
	if len(buf) > 0 {
		list = append(list, string(buf))
	}
	return list, nil
}
