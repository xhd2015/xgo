//go:build go1.21 && !go1.22
// +build go1.21,!go1.22

package patch

import (
	"cmd/compile/internal/types"
)

// different with go1.20
func NewSignature(pkg *types.Pkg, recv *types.Field, tparams, params, results []*types.Field) *types.Type {
	return types.NewSignature(recv, params, results)
}
