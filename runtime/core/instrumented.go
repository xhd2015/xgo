package core

import "github.com/xhd2015/xgo/runtime/internal/runtime"

// Instrumented reports whether this binary has xgo runtime linkage. It is safe
// to call during package initialization and does not depend on whether any
// particular function or package was instrumented.
func Instrumented() bool {
	return runtime.Instrumented()
}
