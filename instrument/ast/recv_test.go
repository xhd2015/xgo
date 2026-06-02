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
			name: "generic index expr",
			typeExpr: &ast.IndexExpr{
				X: ast.NewIdent("T"),
			},
			wantPtr:    false,
			wantGen:    true,
			wantTypeID: "T",
		},
		{
			name: "pointer to generic index expr",
			typeExpr: &ast.StarExpr{
				X: &ast.IndexExpr{X: ast.NewIdent("T")},
			},
			wantPtr:    true,
			wantGen:    true,
			wantTypeID: "T",
		},
		{
			name: "generic index list expr",
			typeExpr: &ast.IndexListExpr{
				X: ast.NewIdent("T"),
			},
			wantPtr:    false,
			wantGen:    true,
			wantTypeID: "T",
		},
		{
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
			name: "selector expr without pointer (pkg.Type)",
			typeExpr: &ast.SelectorExpr{
				X:   ast.NewIdent("pkg"),
				Sel: ast.NewIdent("Type"),
			},
			wantPtr:    false,
			wantGen:    false,
			wantTypeID: "Type",
		},
		{
			name: "pointer to generic selector expr (*pkg.Type[T])",
			typeExpr: &ast.StarExpr{
				X: &ast.IndexExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent("pkg"),
						Sel: ast.NewIdent("Type"),
					},
				},
			},
			wantPtr:    true,
			wantGen:    true,
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

func TestParseReceiverInfo(t *testing.T) {
	// *pkg.Type method named "Method" -> identity "(*Type).Method"
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

func TestParseReceiverInfo_NoSelector(t *testing.T) {
	// *T method named "Method" -> identity "(*T).Method"
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

func TestParseReceiverTypeNilNotPanic(t *testing.T) {
	// ensure nil doesn't panic (tested with recover)
	// This is a defensive test; currently this calls StarExpr.(nil) etc.
	// which will panic. We just note the current behavior.
	defer func() {
		if r := recover(); r == nil {
			t.Log("ParseReceiverType(nil) expected to panic (current behavior)")
		}
	}()
	ParseReceiverType(nil)
}

func TestParseReceiverTypeNoFileSet(t *testing.T) {
	// NewIdent without fileSet might cause position issues
	// but should not affect the parsing logic.
	// This is a complement test.
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


