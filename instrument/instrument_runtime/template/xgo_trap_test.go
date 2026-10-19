package template

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestAppendTrapSkipImpl(t *testing.T) {
	const tmpl = "// comment\n\npackage runtime\n\nimport (\n\t\"unsafe\"\n)\n\nfunc XgoTrap() {}\n"
	for _, c := range []struct {
		minor  int
		atomic string
	}{
		{17, `import __xgo_atomic "runtime/internal/atomic"`},
		{22, `import __xgo_atomic "runtime/internal/atomic"`},
		{23, `import __xgo_atomic "internal/runtime/atomic"`},
		{27, `import __xgo_atomic "internal/runtime/atomic"`},
	} {
		got, err := AppendTrapSkipImpl(&goinfo.GoVersion{Major: 1, Minor: c.minor}, tmpl)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, "package runtime\n\n"+c.atomic+"\n") {
			t.Errorf("go1.%d: atomic import missing after the package clause:\n%s", c.minor, got)
		}
		if strings.Count(got, "func XgoTrapOn(") != 1 || strings.Count(got, "func XgoSetTrapSkip(") != 1 {
			t.Errorf("go1.%d: expected XgoTrapOn and XgoSetTrapSkip once:\n%s", c.minor, got)
		}
		// a template that already defines them is kept as is
		again, err := AppendTrapSkipImpl(&goinfo.GoVersion{Major: 1, Minor: c.minor}, got)
		if err != nil || again != got {
			t.Errorf("go1.%d: appended twice", c.minor)
		}
	}
	if _, err := AppendTrapSkipImpl(&goinfo.GoVersion{Major: 1, Minor: 27}, "package main\n"); err == nil {
		t.Errorf("expected an error without the runtime package clause")
	}
	legacy := tmpl + "\nfunc XgoTrapOn() bool { return true }\n"
	if _, err := AppendTrapSkipImpl(&goinfo.GoVersion{Major: 1, Minor: 27}, legacy); err == nil {
		t.Errorf("expected an error for XgoTrapOn without the flag parameter")
	}
}

func TestTrapEntrySafety(t *testing.T) {
	// Use an old-style template without XgoTrapOn. The installed guard must
	// protect legacy dependencies too, including unguarded variable traps.
	const tmpl = `//go:build ignore

package runtime
func XgoTrap(info interface{}, recvPtr interface{}, args, results []interface{}) (func(), bool) { return nil, false }
func XgoTrapVar(info interface{}, varAddr, res interface{}) {}
func XgoTrapVarPtr(info interface{}, varAddr, res interface{}) {}
`
	for _, minor := range []int{17, 22, 23, 27} {
		version := &goinfo.GoVersion{Major: 1, Minor: minor}
		got, err := InstantiateXgoTrap(tmpl, version)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(token.NewFileSet(), "xgo_trap.go", got, 0)
		if err != nil {
			t.Fatal(err)
		}
		entries := 0
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			switch fn.Name.Name {
			case "XgoTrap", "XgoTrapVar", "XgoTrapVarPtr", "XgoTrapOn":
			default:
				continue
			}
			entries++
			ifStmt, ok := fn.Body.List[0].(*ast.IfStmt)
			if !ok {
				t.Fatalf("go1.%d: %s does not start with a guard", minor, fn.Name.Name)
			}
			neg, ok := ifStmt.Cond.(*ast.UnaryExpr)
			if !ok || neg.Op != token.NOT {
				t.Fatalf("go1.%d: %s guard is not negated", minor, fn.Name.Name)
			}
			call, ok := neg.X.(*ast.CallExpr)
			if !ok {
				t.Fatalf("go1.%d: %s guard is not a call", minor, fn.Name.Name)
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "__xgo_can_trap" {
				t.Fatalf("go1.%d: %s has wrong guard", minor, fn.Name.Name)
			}
			if _, ok := ifStmt.Body.List[0].(*ast.ReturnStmt); !ok {
				t.Fatalf("go1.%d: %s guard does not return", minor, fn.Name.Name)
			}
		}
		if entries != 4 || strings.Count(got, "func __xgo_can_trap()") != 1 {
			t.Fatalf("go1.%d: missing guarded entries or safety helper", minor)
		}
		again, err := guardTrapEntries(got)
		if err != nil || again != got {
			t.Fatalf("go1.%d: guards installed twice: %v", minor, err)
		}
	}
}
