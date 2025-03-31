package init_fn

import (
	"testing"
)

// this test should build without error
// no other assert needed
func TestInitFnShouldNotBeTrapped(t *testing.T) {
	// init() // cannot call init
	init2()

	SomeType(0).init()

	__()
}
