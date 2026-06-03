package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Newline-separated: each declaration on its own line → valid Go
	srcNewline := `package p

import _ "fmt"

var help = ` + "`hello`" + `

func help_xgo_get() string { return help }
func help_xgo_get_addr() *string { return &help }

type __xgo_func_info_0 struct{ Kind int }
var __xgo_register_0 = func(v interface{}) {}
var __xgo_pkg_0 = "test"
func init() { _ = __xgo_register_0 }
func init() { __xgo_register_0(nil) }

func GetHelp() string { return help }
`

	// Semicolon-separated: all on one line, ; between declarations
	// The ";func" at the start creates an empty statement at package level → INVALID
	srcSemicolon := `package p

import _ "fmt"

var help = ` + "`hello`" + `
;func help_xgo_get() string { return help };func help_xgo_get_addr() *string { return &help };type __xgo_func_info_0 struct{ Kind int };var __xgo_register_0 = func(v interface{}) {};var __xgo_pkg_0 = "test";func init() { _ = __xgo_register_0 };func init() { __xgo_register_0(nil) }func GetHelp() string { return help }
`

	// Test A: newline version → must pass
	fmt.Println("=== Test A: newline-separated (expected: PASS) ===")
	testCompile(srcNewline, true)

	// Test B: semicolon version → must fail
	fmt.Println("=== Test B: semicolon-separated (expected: FAIL) ===")
	testCompile(srcSemicolon, false)

	fmt.Println("=== ALL TESTS PASSED ===")
}

func testCompile(src string, expectPass bool) {
	tmpDir, err := os.MkdirTemp("", "xgo-semi-*")
	if err != nil {
		fatalf("create temp dir: %v", err)
	}
	// Keep dir for inspection
	fmt.Printf("  dir: %s\n", tmpDir)

	goMod := "module test\n\ngo 1.21\n"
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(src), 0644)

	cmd := exec.Command("go", "build", ".")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()

	if err != nil {
		if expectPass {
			fmt.Printf("  UNEXPECTED FAILURE:\n%s\n", string(out))
			fatalf("expected compilation to PASS but got ERROR")
		}
		fmt.Printf("  FAIL (expected): %s\n", firstLine(string(out)))
	} else {
		if !expectPass {
			fatalf("expected compilation to FAIL but got PASS")
		}
		fmt.Println("  PASS")
	}
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}
