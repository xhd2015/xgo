//go:build go1.20 && !go1.21
// +build go1.20,!go1.21

package patch

import "cmd/compile/internal/types"

// different with go1.20
func NewSignature(pkg *types.Pkg, recv *types.Field, tparams, params, results []*types.Field) *types.Type {
	return types.NewSignature(pkg, recv, tparams, params, results)
}
