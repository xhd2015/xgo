package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

func main() {
	// Build source with \r bytes INSIDE the raw string.
	// Git's core.autocrlf=true injects \r into all \n,
	// including \n inside backtick-delimited raw string literals.
	// Go's scanner strips \r from BasicLit.Value but byte offsets
	// in token.FileSet include them, causing End() to be short.
	src := "package main\r\n\r\nimport \"fmt\"\r\n\r\nconst help = `\r\nUsage:\r\n  -func match\r\n`\r\n\r\nfunc main() {\r\n\tfmt.Print(help)\r\n}\r\n"

	if !strings.Contains(src, "\r") {
		fatalf("source does not contain \\r bytes — test invalid")
	}

	// --- Test A: old approach (decl.End() + ; separator) ---
	fmt.Println("=== Test A: OLD approach (decl.End() + ;) ===")
	fmt.Println("    Expected: stub lands INSIDE raw string (bad)")
	testApproach(src, "old", func(file *ast.File, declIndex int) token.Pos {
		return file.Decls[declIndex].End()
	}, ";", true)

	// --- Test B: new approach (safe end + \n separator) ---
	fmt.Println("\n=== Test B: NEW approach (safeEnd + \\n) ===")
	fmt.Println("    Expected: stub lands OUTSIDE raw string (good)")
	testApproach(src, "new", func(file *ast.File, declIndex int) token.Pos {
		return safeEnd(file, declIndex)
	}, "\n", false)

	fmt.Println("\n=== ALL TESTS PASSED ===")
}

// testApproach creates a temp dir, writes the source, edits it, and checks
// whether the inserted stub is inside the raw string (expectInside=true means
// we expect the stub to be buried inside the raw string — which is the bug).
func testApproach(src, name string, posFn func(*ast.File, int) token.Pos, sep string, expectInside bool) {
	tmpDir, err := os.MkdirTemp("", "xgo-cr-"+name+"-*")
	if err != nil {
		fatalf("[%s] create temp dir: %v", name, err)
	}
	fmt.Printf("[%s] test dir: %s\n", name, tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		fatalf("[%s] write go.mod: %v", name, err)
	}

	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(src), 0644); err != nil {
		fatalf("[%s] write main.go: %v", name, err)
	}

	// Baseline: original compiles and runs
	fmt.Printf("[%s] baseline go run: ", name)
	if out, err := runGo(tmpDir, "."); err != nil {
		fatalf("[%s] baseline failed: %v\n%s", name, err, out)
	}
	fmt.Println("OK")

	// Parse
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, srcPath, nil, parser.ParseComments)
	if err != nil {
		fatalf("[%s] parse: %v", name, err)
	}

	declIndex := firstNonImportGenDecl(file)
	if declIndex < 0 {
		fatalf("[%s] no non-import GenDecl found", name)
	}

	gd := file.Decls[declIndex].(*ast.GenDecl)
	lit := gd.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit)
	endOffset := fset.Position(lit.End()).Offset
	nextOffset := fset.Position(file.Decls[declIndex+1].Pos()).Offset
	fmt.Printf("[%s] BasicLit.End() offset=%d  nextDecl.Pos() offset=%d  (gap=%d)\n",
		name, endOffset, nextOffset, nextOffset-endOffset)

	pos := posFn(file, declIndex)
	insertOffset := fset.Position(pos).Offset
	fmt.Printf("[%s] insert position offset=%d\n", name, insertOffset)

	// Apply edit
	edit := goedit.New(fset, src)
	stubFunc := "func help_xgo_get() string { return help }"

	if sep == ";" {
		edit.Insert(pos, ";")
		edit.Insert(pos, stubFunc)
	} else {
		edit.Insert(pos, "\n"+stubFunc+"\n")
	}

	edited := edit.String()

	// Write and check
	if err := os.Remove(srcPath); err != nil {
		fatalf("[%s] remove original: %v", name, err)
	}
	if err := os.WriteFile(srcPath, []byte(edited), 0644); err != nil {
		fatalf("[%s] write edited: %v", name, err)
	}

	content, _ := os.ReadFile(srcPath)
	inside := inRawStr(content, stubFunc)

	if inside {
		fmt.Printf("[%s] stub is INSIDE raw string — invisible to compiler\n", name)
		// Show a snippet
		if idx := bytes.Index(content, []byte(stubFunc)); idx >= 0 {
			s := idx - 30
			if s < 0 {
				s = 0
			}
			e := idx + len(stubFunc) + 30
			if e > len(content) {
				e = len(content)
			}
			fmt.Printf("[%s] snippet: %q\n", name, string(content[s:e]))
		}
	} else {
		fmt.Printf("[%s] stub is OUTSIDE raw string — OK\n", name)
	}

	if expectInside && !inside {
		fatalf("[%s] expected stub inside raw string, but it's outside — test may be invalid", name)
	}
	if !expectInside && inside {
		fatalf("[%s] expected stub outside raw string, but it's inside — BUG", name)
	}

	// Also verify go run still works (for new approach it should)
	fmt.Printf("[%s] edited go run: ", name)
	if out, err := runGo(tmpDir, "."); err != nil {
		fmt.Printf("FAIL\n%s", out)
		if !expectInside {
			fatalf("[%s] edited file should compile but didn't", name)
		}
	} else {
		fmt.Println("OK")
	}
}

func firstNonImportGenDecl(file *ast.File) int {
	for i, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok != token.IMPORT {
			return i
		}
	}
	return -1
}

func safeEnd(file *ast.File, declIndex int) token.Pos {
	if declIndex+1 < len(file.Decls) {
		nextDecl := file.Decls[declIndex+1]
		if !hasGoDirective(nextDecl) {
			return nextDecl.Pos()
		}
	}
	return file.Decls[declIndex].End()
}

func hasGoDirective(decl ast.Decl) bool {
	var doc *ast.CommentGroup
	switch d := decl.(type) {
	case *ast.GenDecl:
		doc = d.Doc
	case *ast.FuncDecl:
		doc = d.Doc
	}
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.HasPrefix(c.Text, "//go:") {
			return true
		}
	}
	return false
}

func inRawStr(src []byte, snippet string) bool {
	idx := bytes.Index(src, []byte(snippet))
	if idx < 0 {
		return false
	}
	inRaw := false
	for i := 0; i < idx; i++ {
		if src[i] == '`' {
			inRaw = !inRaw
		}
	}
	return inRaw
}

func runGo(dir string, pkg string) (string, error) {
	cmd := exec.Command("go", "run", pkg)
	cmd.Dir = dir
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return stderr.String() + "\n" + stdout.String(), err
	}
	return stdout.String(), nil
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}
