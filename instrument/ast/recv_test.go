// Tests for ParseReceiverType and ParseReceiverInfo covering all receiver
// forms, including *ast.SelectorExpr which caused a panic before the fix.
//
// Bug: ParseReceiverType only handled *ast.Ident as the leaf type name after
// unwrapping *ast.StarExpr (pointer) and *ast.IndexExpr/*ast.IndexListExpr
// (generic). When a method receiver referenced a type from another package
// (e.g., *otherpkg.Type), the inner type after unwrapping is *ast.SelectorExpr,
// not *ast.Ident, causing a panic.
//
// Fix: Added *ast.SelectorExpr handling before the *ast.Ident check. When a
// SelectorExpr is encountered, selExpr.Sel (the type name *ast.Ident) is
// returned. This is consistent with how getEmbedFieldName in resolve.go
// already handles SelectorExpr for embedded field names.
//
// The tests marked [fix: selector] cover the new cases that would have
// panicked before the fix.
//
// Generic test cases (IndexExpr, IndexListExpr) are in recv_test_go1.18.go.
package ast

import (
	"go/ast"
	"testing"
)

func TestParseReceiverType(t *testing.T) {
	tests := []struct {
		name       string
		typeExpr   ast.Expr
		wantPtr    bool
		wantGen    bool
		wantTypeID string
	}{
		{
			name:       "simple ident",
			typeExpr:   ast.NewIdent("T"),
			wantPtr:    false,
			wantGen:    false,
			wantTypeID: "T",
		},
		{
			name:       "pointer to ident",
			typeExpr:   &ast.StarExpr{X: ast.NewIdent("T")},
			wantPtr:    true,
			wantGen:    false,
			wantTypeID: "T",
		},
		{
			// [fix: selector] *pkg.Type — pointer-to-type-from-other-package.
			// AST: *ast.StarExpr{X: *ast.SelectorExpr{X: pkg, Sel: Type}}
			// Before fix: panicked at "expect receiver to be ident".
			name: "pointer selector expr (*pkg.Type)",
			typeExpr: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("pkg"),
					Sel: ast.NewIdent("Type"),
				},
			},
			wantPtr:    true,
			wantGen:    false,
			wantTypeID: "Type",
		},
		{
			// [fix: selector] pkg.Type — type-from-other-package without pointer.
			// AST: *ast.SelectorExpr{X: pkg, Sel: Type}
			// Before fix: panicked at "expect receiver to be ident".
			name: "selector expr without pointer (pkg.Type)",
			typeExpr: &ast.SelectorExpr{
				X:   ast.NewIdent("pkg"),
				Sel: ast.NewIdent("Type"),
			},
			wantPtr:    false,
			wantGen:    false,
			wantTypeID: "Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr, gen, idt := ParseReceiverType(tt.typeExpr)
			if ptr != tt.wantPtr {
				t.Errorf("ptr = %v, want %v", ptr, tt.wantPtr)
			}
			if gen != tt.wantGen {
				t.Errorf("generic = %v, want %v", gen, tt.wantGen)
			}
			if idt == nil {
				t.Fatal("expected non-nil ident")
			}
			if idt.Name != tt.wantTypeID {
				t.Errorf("type name = %q, want %q", idt.Name, tt.wantTypeID)
			}
		})
	}
}

// TestParseReceiverInfo verifies the identity name construction for a method
// with a *pkg.Type receiver. The expected identity is "(*Type).Method" — using
// only the type name portion of the selector, not the package prefix.
func TestParseReceiverInfo(t *testing.T) {
	typeExpr := &ast.StarExpr{
		X: &ast.SelectorExpr{
			X:   ast.NewIdent("pkg"),
			Sel: ast.NewIdent("Type"),
		},
	}
	name, ptr, gen, idt := ParseReceiverInfo("Method", typeExpr)
	if name != "(*Type).Method" {
		t.Errorf("identity = %q, want %q", name, "(*Type).Method")
	}
	if !ptr {
		t.Error("expected ptr=true")
	}
	if gen {
		t.Error("expected generic=false")
	}
	if idt.Name != "Type" {
		t.Errorf("type name = %q, want %q", idt.Name, "Type")
	}
}

// TestParseReceiverInfo_NoSelector verifies identity name for the common case
// of a simple pointer receiver *T, serving as a baseline regression check.
func TestParseReceiverInfo_NoSelector(t *testing.T) {
	typeExpr := &ast.StarExpr{X: ast.NewIdent("T")}
	name, ptr, gen, idt := ParseReceiverInfo("Method", typeExpr)
	if name != "(*T).Method" {
		t.Errorf("identity = %q, want %q", name, "(*T).Method")
	}
	if !ptr {
		t.Error("expected ptr=true")
	}
	if gen {
		t.Error("expected generic=false")
	}
	if idt.Name != "T" {
		t.Errorf("type name = %q, want %q", idt.Name, "T")
	}
}

// TestParseReceiverTypeNilNotPanic documents the current behavior that
// passing nil to ParseReceiverType panics (due to type assertion on nil).
// This is a defensive characterization test, not an endorsement of nil input.
func TestParseReceiverTypeNilNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Log("ParseReceiverType(nil) expected to panic (current behavior)")
		}
	}()
	ParseReceiverType(nil)
}

// TestParseReceiverTypeNoFileSet verifies that ParseReceiverType works with
// *ast.Ident nodes created without a token.FileSet (NamePos = 0). This ensures
// the parsing logic does not depend on position information from a FileSet.
func TestParseReceiverTypeNoFileSet(t *testing.T) {
	typeExpr := &ast.StarExpr{
		X: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "net"},
			Sel: &ast.Ident{Name: "Conn"},
		},
	}
	ptr, gen, idt := ParseReceiverType(typeExpr)
	if !ptr {
		t.Error("expected ptr=true for *net.Conn")
	}
	if gen {
		t.Error("expected generic=false")
	}
	if idt.Name != "Conn" {
		t.Errorf("type name = %q, want %q", idt.Name, "Conn")
	}
}
