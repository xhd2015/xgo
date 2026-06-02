//go:build go1.18
// +build go1.18

// Tests for ParseReceiverType covering generic receiver forms that require
// Go 1.18+ AST types (*ast.IndexExpr and *ast.IndexListExpr).
//
// These are split from recv_test.go to avoid build failures on pre-1.18 Go.

package ast

import (
	"go/ast"
	"testing"
)

func TestParseReceiverType_Generic(t *testing.T) {
	tests := []struct {
		name       string
		typeExpr   ast.Expr
		wantPtr    bool
		wantGen    bool
		wantTypeID string
	}{
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
			// [fix: selector] *pkg.Type[T] — pointer-to-generic-type-from-other-package.
			// AST: *ast.StarExpr{X: *ast.IndexExpr{X: *ast.SelectorExpr{X: pkg, Sel: Type}}}
			// Before fix: panicked after unwrapping both StarExpr and IndexExpr.
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
