// patch reflect package
// NOTE: not used currently
package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/transform"
)

func addReflectFunctions(goroot string, goVersion *goinfo.GoVersion, xgoSrc string) error {
	dstFile := filepath.Join(goroot, "src", "reflect", "xgo_reflect.go")
	content, err := readXgoSrc(xgoSrc, []string{"trap_runtime", "xgo_reflect.go"})
	if err != nil {
		return err
	}

	content, err = replaceBuildIgnore(content)
	if err != nil {
		return fmt.Errorf("file %s: %w", filepath.Base(dstFile), err)
	}

	valCode, err := transformReflectValue(filepath.Join(goroot, "src", "reflect", "value.go"))
	if err != nil {
		return fmt.Errorf("transforming reflect/value.go: %w", err)
	}
	typeCode, err := transformReflectType(filepath.Join(goroot, "src", "reflect", "type.go"))
	if err != nil {
		return fmt.Errorf("transforming reflect/type.go: %w", err)
	}

	// fmt.Printf("typCode: %s\n", typeCode)

	// concat all code
	content = bytes.Join([][]byte{content, []byte(valCode), []byte(typeCode)}, []byte("\n"))
	return os.WriteFile(dstFile, content, 0755)
}

const xgoGetAllMethodByName = "__xgo_get_all_method_by_name"

func transformReflectValue(reflectValueFile string) (string, error) {
	file, err := transform.Parse(reflectValueFile)
	if err != nil {
		return "", err
	}

	fnDecl := file.GetMethodDecl("Value", "MethodByName")
	if fnDecl == nil {
		return "", fmt.Errorf("cannot find Value.MethodByName")
	}

	code, err := replaceIdent(file, fnDecl, xgoGetAllMethodByName, func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		switch idt.Name {
		case "MethodByName":
			return idt, xgoGetAllMethodByName
		case "Method": // method by index
			return idt, "__xgo_get_all_method_index"
		}

		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing MethodByName: %w", err)
	}

	methodDecl := file.GetMethodDecl("Value", "Method") // method by index
	if methodDecl == nil {
		return "", fmt.Errorf("cannot find Value.Method")
	}
	code2, err := replaceIdent(file, methodDecl, "__xgo_get_all_method_index", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		switch idt.Name {
		case "NumMethod": // method by index
			return idt, "__xgo_get_all_method_num"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing Method: %w", err)
	}

	codef := strings.Join([]string{code, code2}, "\n")
	return codef, nil
}

func transformReflectType(reflectTypeFile string) (string, error) {
	file, err := transform.Parse(reflectTypeFile)
	if err != nil {
		return "", err
	}
	fnDecl := file.GetMethodDecl("rtype", "MethodByName")
	if fnDecl == nil {
		return "", fmt.Errorf("cannot find rtype.MethodByName")
	}
	m0, err := replaceIdent(file, fnDecl, xgoGetAllMethodByName, func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "ExportedMethods" {
			return idt, "Methods"
		} else if idt.Name == "Method" {
			return idt, "__xgo_get_all_method_index"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing ExportedMethods: %w", err)
	}

	fnDecl2 := file.GetMethodDecl("rtype", "exportedMethods")
	if fnDecl2 == nil {
		return "", fmt.Errorf("cannot find rtype.exportedMethods")
	}

	m1, err := replaceIdent(file, fnDecl2, "__xgo_all_methods", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "ExportedMethods" {
			return idt, "Methods"
		}
		return nil, ""
	})
	if err != nil {
		return "", err
	}

	methodDecl := file.GetMethodDecl("rtype", "Method")
	if methodDecl == nil {
		return "", fmt.Errorf("cannot find rtype.Method")
	}
	m2, err := replaceIdent(file, methodDecl, "__xgo_get_all_method_index", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "exportedMethods" {
			return idt, "__xgo_all_methods"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing Method: %w", err)
	}

	numA := file.GetMethodDecl("rtype", "NumMethod")
	if numA == nil {
		return "", fmt.Errorf("cannot find rtype.NumMethod")
	}
	m3, err := replaceIdent(file, numA, "__xgo_get_all_method_num", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "exportedMethods" {
			return idt, "__xgo_all_methods"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing Method: %w", err)
	}
	code := strings.Join([]string{m0, m1, m2, m3}, "\n")
	return code, nil
}

func replaceIdent(file *transform.File, fnDecl *ast.FuncDecl, replaceFuncName string, identReplacer func(n ast.Node) (*ast.Ident, string)) (string, error) {
	type replaceIdent struct {
		idt *ast.Ident
		rep string
	}
	var replaceIdents []replaceIdent
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		if n == nil {
			// post action
			return false
		}
		idt, replace := identReplacer(n)
		if idt != nil {
			replaceIdents = append(replaceIdents, replaceIdent{
				idt: idt,
				rep: replace,
			})
		}
		return true
	})
	if len(replaceIdents) == 0 {
		return "", errors.New("no replace found")
	}
	if replaceFuncName != "" {
		// replace the name
		replaceIdents = append(replaceIdents, replaceIdent{
			idt: fnDecl.Name,
			rep: replaceFuncName,
		})
	}
	// find assignment to x
	sort.Slice(replaceIdents, func(i, j int) bool {
		a := replaceIdents[i].idt
		b := replaceIdents[j].idt
		return file.Fset.Position(a.Pos()).Offset < file.Fset.Position(b.Pos()).Offset
	})

	// replace
	n := len(replaceIdents)
	baseOffset := file.Fset.Position(fnDecl.Pos()).Offset

	code := file.GetCode(fnDecl)
	for i := n - 1; i >= 0; i-- {
		rp := replaceIdents[i]
		offset := file.Fset.Position(rp.idt.Pos()).Offset - baseOffset

		var buf bytes.Buffer
		buf.Grow(len(code))
		buf.Write(code[:offset])
		buf.WriteString(rp.rep)
		buf.Write(code[offset+len(rp.idt.Name):])

		code = buf.Bytes()
		// NOTE: don't use slice append, content will be override
		if false {
			newCode := append(code[:offset:offset], []byte(rp.rep)...)
			newCode = append(newCode, code[offset+len(rp.idt.Name):]...)
			code = newCode
		}
	}
	return string(code), nil
}
