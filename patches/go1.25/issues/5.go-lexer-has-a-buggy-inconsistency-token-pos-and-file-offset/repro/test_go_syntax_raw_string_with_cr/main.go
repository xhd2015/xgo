package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func main() {
	// LF-only raw string
	srcLF := "package p\n\nvar help = `\nhello\n`\n"
	// CRLF raw string (simulating git core.autocrlf on Windows)
	srcCRLF := "package p\r\n\r\nvar help = `\r\nhello\r\n`\r\n"

	fmt.Println("=== Raw string, LF-only ===")
	analyzeRaw(srcLF, "LF")

	fmt.Println("\n=== Raw string, CRLF ===")
	analyzeRaw(srcCRLF, "CRLF")

	// Interpreted string with \r inside (NOT \r\n, since \n would be parse error)
	srcInterpretedLF := "package p\n\nvar help = \"A\x0dB\"\n"
	srcInterpretedCRLF := "package p\r\n\r\nvar help = \"A\x0dB\"\r\n"

	fmt.Println("\n=== Interpreted string with \\r, LF ===")
	analyzeInterpreted(srcInterpretedLF, "LF, \\r in \"...\"")

	fmt.Println("\n=== Interpreted string with \\r, CRLF ===")
	analyzeInterpreted(srcInterpretedCRLF, "CRLF, \\r in \"...\"")

	fmt.Println("\n=== CONCLUSION ===")
	fmt.Println("Raw string:   BasicLit.End() wrong by \\r count inside backticks")
	fmt.Println("Interpreted:  BasicLit.End() correct — \\r is included in len(Value)")
}

func analyzeRaw(src string, label string) {
	crCount := strings.Count(src, "\r")
	fmt.Printf("[%s] source: %d bytes, %d \\r\n", label, len(src), crCount)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		fmt.Printf("  parse error: %v\n", err)
		return
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs := spec.(*ast.ValueSpec)
			lit := vs.Values[0].(*ast.BasicLit)

			valuePosOffset := fset.Position(lit.ValuePos).Offset
			endOffset := fset.Position(lit.End()).Offset
			closeBacktick := findDelim(src, valuePosOffset+1, '`')
			realEndOffset := closeBacktick + 1

			fmt.Printf("  Value=%q  ValuePos=%d  len(Value)=%d  End()=%d\n",
				lit.Value, valuePosOffset, len(lit.Value), endOffset)
			fmt.Printf("  real closing ` at %d  real End()=%d  gap=%d\n",
				closeBacktick, realEndOffset, realEndOffset-endOffset)
		}
	}
}

func analyzeInterpreted(src string, label string) {
	crCount := strings.Count(src, "\r")
	fmt.Printf("[%s] source: %d bytes, %d \\r\n", label, len(src), crCount)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		fmt.Printf("  parse error: %v\n", err)
		return
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs := spec.(*ast.ValueSpec)
			lit := vs.Values[0].(*ast.BasicLit)

			valuePosOffset := fset.Position(lit.ValuePos).Offset
			endOffset := fset.Position(lit.End()).Offset
			closeQuote := findDelim(src, valuePosOffset+1, '"')
			realEndOffset := closeQuote + 1

			fmt.Printf("  Value=%q  ValuePos=%d  len(Value)=%d  End()=%d\n",
				lit.Value, valuePosOffset, len(lit.Value), endOffset)
			fmt.Printf("  real closing \" at %d  real End()=%d  gap=%d\n",
				closeQuote, realEndOffset, realEndOffset-endOffset)
		}
	}
}

func findDelim(src string, start int, delim byte) int {
	for i := start; i < len(src); i++ {
		if src[i] == delim {
			return i
		}
	}
	return -1
}
