package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func main() {
	// LF-only source: a raw string that spans 3 lines
	srcLF := "package p\n\nconst help = `\nUsage:\n  -func match\n`\n"

	// CRLF source: same, with \r\n line endings (simulates git core.autocrlf on Windows)
	srcCRLF := "package p\r\n\r\nconst help = `\r\nUsage:\r\n  -func match\r\n`\r\n"

	fmt.Println("=== LF (\\n only) ===")
	demo(srcLF)

	fmt.Println("\n=== CRLF (\\r\\n) ===")
	demo(srcCRLF)

	fmt.Println("\n=== SUMMARY ===")
	fmt.Println("Go's scanner.go:667-710 strips \\r from BasicLit.Value (via stripCR)")
	fmt.Println("but BasicLit.ValuePos tracks real byte offsets (including \\r)")
	fmt.Println("=> BasicLit.End() = ValuePos + len(Value) is short by \\r count")
	fmt.Println("")
	fmt.Println("Commit: fb6ffd8f787f (Dec 15, 2011, Robert Griesemer)")
	fmt.Println("Files:  go/scanner/scanner.go:667-682 (stripCR)")
	fmt.Println("        go/scanner/scanner.go:684-710 (scanRawString)")
}

func demo(src string) {
	crCount := strings.Count(src, "\r")
	fmt.Printf("source: %d bytes, %d \\r bytes\n", len(src), crCount)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		fmt.Printf("  parse error: %v\n", err)
		return
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs := spec.(*ast.ValueSpec)
			lit := vs.Values[0].(*ast.BasicLit)

			// AST-computed position
			astEndOff := fset.Position(lit.End()).Offset

			// Real closing backtick by scanning source bytes
			valOff := fset.Position(lit.ValuePos).Offset
			closeBTOff := -1
			for i := valOff + 1; i < len(src); i++ {
				if src[i] == '`' {
					closeBTOff = i
					break
				}
			}
			realEndOff := closeBTOff + 1
			gap := realEndOff - astEndOff

			fmt.Printf("  Value         = %q\n", lit.Value)
			fmt.Printf("  len(Value)    = %d\n", len(lit.Value))
			fmt.Printf("  ValuePos      = offset %d\n", valOff)
			fmt.Printf("  End()         = offset %d (ValuePos + len)\n", astEndOff)
			fmt.Printf("  real closing `= offset %d\n", closeBTOff)
			fmt.Printf("  real End()    = offset %d\n", realEndOff)

			if gap == 0 {
				fmt.Println("  ✓ End() CORRECT (gap=0)")
			} else {
				fmt.Printf("  ✗ End() WRONG by %d bytes\n", gap)
				fmt.Printf("    bytes between End() and closing `: %q\n", src[astEndOff:realEndOff])
			}
		}
	}
}
