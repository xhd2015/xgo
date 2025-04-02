//go:build go1.18
// +build go1.18

package main

import (
	"go/ast"

	"github.com/xhd2015/xgo/support/fileutil"
)

func patchJSONPretty(settingsFile string, fn func(settings *map[string]interface{}) error) error {
	return fileutil.PatchJSONPretty(settingsFile, fn)
}

func isGeneric(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.TypeParams == nil || len(funcDecl.Type.TypeParams.List) == 0 {
		return false
	}
	return true
}
