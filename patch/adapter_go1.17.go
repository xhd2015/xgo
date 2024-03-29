//go:build go1.17 && !go1.18
// +build go1.17,!go1.18

package patch

import "cmd/compile/internal/ir"

const goMajor = 1
const goMinor = 17

func isClosureWrapperForGeneric(fn *ir.Func) bool {
	return false
}
