package instrument_var

import (
	"fmt"
	"go/ast"
)

func getFuncDeclNamesNoBlank(recv *ast.FieldList, funcType *ast.FuncType) []string {
	var names []string
	recvNames := getFieldNames(recv)
	if len(recvNames) > 1 {
		panic(fmt.Errorf("want at most one receiver, got %d", len(recvNames)))
	}

	if len(recvNames) > 0 && !isBlankName(recvNames[0]) {
		names = append(names, recvNames[0])
	}
	paramNames := getFieldNames(funcType.Params)
	for _, name := range paramNames {
		if !isBlankName(name) {
			names = append(names, name)
		}
	}
	resultNames := getFieldNames(funcType.Results)
	for _, name := range resultNames {
		if !isBlankName(name) {
			names = append(names, name)
		}
	}
	return names
}

func isBlankName(name string) bool {
	return name == "" || name == "_"
}

func getFieldNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	names := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		names = append(names, getFieldName(field))
	}
	return names
}
func getFieldName(f *ast.Field) string {
	if f == nil || f.Names == nil {
		return ""
	}
	return f.Names[0].Name
}
