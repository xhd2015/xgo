package repro

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/edit/goedit"
)

// LF-only source with a raw string literal. Converted to CRLF at runtime
// to simulate git core.autocrlf=true on Windows injecting \r bytes.
const srcLF = "package p\n\nconst help = `\nUsage:\n  -func match\n`\n\nfunc Hello() {}\n"

const stub = "var __xgo_trap_0 = func(){}"

func TestCRLF_rawString_stubInside(t *testing.T) {
	src := strings.ReplaceAll(srcLF, "\n", "\r\n")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	// ── Buggy: insert at genDecl.End() ─────────────────────────────────
	edit := goedit.NewWithBytes(fset, []byte(src))
	patch.AddCode(edit, f.Decls[0].End(), stub)
	buggyResult := edit.String()

	// Verify it's still valid Go source (even if semantically broken)
	_, err = parser.ParseFile(token.NewFileSet(), "", buggyResult, parser.ParseComments)
	if err != nil {
		t.Logf("BUGGY parse error (expected): %v", err)
	}

	// Check stub location relative to raw string boundary
	rawEnd := findRawStringEnd(buggyResult)
	stubAt := strings.Index(buggyResult, stub)
	if stubAt < 0 {
		t.Fatal("stub not found in buggy output")
	}
	if stubAt < rawEnd {
		t.Logf("BUGGY: stub at byte %d, raw string ends at byte %d → stub INSIDE raw string", stubAt, rawEnd)
	} else {
		t.Error("BUGGY: expected stub inside raw string, but it is outside")
	}

	// ── Fixed: insert at GenDeclSafeEnd() ──────────────────────────────
	fset2 := token.NewFileSet()
	f2, err := parser.ParseFile(fset2, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	idx := firstNonImportGenDeclIndex(f2)
	safePos := patch.GenDeclSafeEnd(f2, idx)

	edit2 := goedit.NewWithBytes(fset2, []byte(src))
	patch.AddCode(edit2, safePos, stub)
	fixedResult := edit2.String()

	rawEnd2 := findRawStringEnd(fixedResult)
	stubAt2 := strings.Index(fixedResult, stub)
	if stubAt2 < 0 {
		t.Fatal("stub not found in fixed output")
	}
	if stubAt2 > rawEnd2 {
		t.Logf("FIXED: stub at byte %d, raw string ends at byte %d → stub OUTSIDE raw string", stubAt2, rawEnd2)
	} else {
		t.Error("FIXED: stub should be outside raw string, but it is inside")
	}
}

func TestLF_rawString_stubOutside(t *testing.T) {
	// LF source (no \r) — genDecl.End() should be correct on LF-only files.
	// This is the control: verifies genDecl.End() works when no \r is present.
	src := srcLF

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	edit := goedit.NewWithBytes(fset, []byte(src))
	patch.AddCode(edit, f.Decls[0].End(), stub)
	result := edit.String()

	rawEnd := findRawStringEnd(result)
	stubAt := strings.Index(result, stub)
	if stubAt < 0 {
		t.Fatal("stub not found in output")
	}
	if stubAt > rawEnd {
		t.Logf("LF: stub at byte %d, raw string ends at byte %d → stub OUTSIDE raw string", stubAt, rawEnd)
	} else {
		t.Error("LF: expected stub outside raw string, but it is inside")
	}
}

func findRawStringEnd(s string) int {
	// Find the closing backtick of a raw string literal.
	// Assumes exactly one raw string in the source.
	start := strings.IndexByte(s, '`')
	if start < 0 {
		return -1
	}
	for i := start + 1; i < len(s); i++ {
		if s[i] == '`' {
			return i
		}
	}
	return -1
}

func firstNonImportGenDeclIndex(f *ast.File) int {
	for i, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok != token.IMPORT {
			return i
		}
	}
	return -1
}
