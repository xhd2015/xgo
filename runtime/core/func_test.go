package core

import (
	"strings"
	"testing"
)

// go test -run TestParseFuncName -v ./core
func TestParseFuncName(t *testing.T) {
	var testCases = []struct {
		FullName    string
		PkgPath     string
		RecvName    string
		RecvPtr     bool
		TypeGeneric string
		FuncGeneric string
		FuncName    string
	}{
		{"fmt.Printf", "fmt", "", false, "", "", "Printf"},
		{"os.File.Read", "os", "File", false, "", "", "Read"},
		{"os.(*File).Read", "os", "File", true, "", "", "Read"},
		{"github.com/xhd2015/xgo.(*File).Read", "github.com/xhd2015/xgo", "File", true, "", "", "Read"},
		{"github.com/xhd2015/xgo.1.(*File).Read", "github.com/xhd2015/xgo.1", "File", true, "", "", "Read"},     // pkg path with dot
		{"github.com/xhd2015/xgo.(*File[int]).Read", "github.com/xhd2015/xgo", "File", true, "int", "", "Read"}, // generic
		{"github.com/xhd2015/xgo.(*File[int]).Read[string]", "github.com/xhd2015/xgo", "File", true, "int", "string", "Read"},
		{"github.com/xhd2015/xgo.Watch", "github.com/xhd2015/xgo", "", false, "", "", "Watch"},
		{"github.com/xhd2015/xgo.Watch[int]", "github.com/xhd2015/xgo", "", false, "", "int", "Watch"},
		{
			"github.com/xhd2015/xgo/runtime/test/debug.GenericSt[github.com/xhd2015/xgo/runtime/test/debug.Inner].GetData",
			"github.com/xhd2015/xgo/runtime/test/debug",
			"GenericSt",
			false,
			"github.com/xhd2015/xgo/runtime/test/debug.Inner",
			"",
			"GetData",
		},
	}
	for _, tt := range testCases {
		name := strings.ReplaceAll(tt.FullName, "/", "_")
		name = strings.ReplaceAll(name, "[", "_")
		name = strings.ReplaceAll(name, "]", "_")
		t.Run(name, func(t *testing.T) {
			// t.Logf("parse: %s", testCase.FullName)
			pkgPath, recvName, recvPtr, typeGeneric, funcGeneric, funcName := ParseFuncName(tt.FullName)
			if pkgPath != tt.PkgPath {
				t.Fatalf("expect PkgPath to be %s, actual: %s", tt.PkgPath, pkgPath)
			}
			if recvName != tt.RecvName {
				t.Fatalf("expect RecvName to be %s, actual: %s", tt.RecvName, recvName)
			}
			if recvPtr != tt.RecvPtr {
				t.Fatalf("expect RecvPtr to be %v, actual: %v", tt.RecvPtr, recvPtr)
			}
			if typeGeneric != tt.TypeGeneric {
				t.Fatalf("expect TypeGeneric to be %s, actual: %s", tt.TypeGeneric, typeGeneric)
			}
			if funcGeneric != tt.FuncGeneric {
				t.Fatalf("expect FuncGeneric to be %s, actual: %s", tt.FuncGeneric, funcGeneric)
			}
			if funcName != tt.FuncName {
				t.Fatalf("expect FuncName to be %s, actual: %s", tt.FuncName, funcName)
			}
		})
	}
}
