package core

import "testing"

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
	}
	for _, testCase := range testCases {
		// t.Logf("parse: %s", testCase.FullName)
		pkgPath, recvName, recvPtr, typeGeneric, funcGeneric, funcName := ParseFuncName(testCase.FullName)
		if pkgPath != testCase.PkgPath {
			t.Fatalf("expect PkgPath to be %s, actual: %s", testCase.PkgPath, pkgPath)
		}
		if recvName != testCase.RecvName {
			t.Fatalf("expect RecvName to be %s, actual: %s", testCase.RecvName, recvName)
		}
		if recvPtr != testCase.RecvPtr {
			t.Fatalf("expect RecvPtr to be %v, actual: %v", testCase.RecvPtr, recvPtr)
		}
		if typeGeneric != testCase.TypeGeneric {
			t.Fatalf("expect TypeGeneric to be %s, actual: %s", testCase.TypeGeneric, typeGeneric)
		}
		if funcGeneric != testCase.FuncGeneric {
			t.Fatalf("expect FuncGeneric to be %s, actual: %s", testCase.FuncGeneric, funcGeneric)
		}
		if funcName != testCase.FuncName {
			t.Fatalf("expect FuncName to be %s, actual: %s", testCase.FuncName, funcName)
		}
	}
}
