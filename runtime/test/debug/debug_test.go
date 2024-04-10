// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/trap"
)

func ToString[T any](v T) string {
	return fmt.Sprint(v)
}

func TestNakedTrapShouldAvoidRecursive(t *testing.T) {
	trap.InspectPC(ToString[int])
	// _, fnInfo, funcPC, trappingPC := trap.InspectPC(ToString[int])
	// _, fnInfoStr, funcPCStr, trappingPCStr := trap.InspectPC(ToString[string])
}
