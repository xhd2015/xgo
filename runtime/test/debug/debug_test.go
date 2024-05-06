// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/mock_var/sub"
)

var c sub.Mapping = sub.Mapping{
	1: "hello",
}

func TestThirdPartyTypeMethodVar(t *testing.T) {
	mock.Patch(&c, func() sub.Mapping {
		return sub.Mapping{
			1: "mock",
		}
	})
	txt := c.Get(1)
	if txt != "mock" {
		t.Fatalf("expect c[1] to be %s, actual: %s", "mock", txt)
	}
}
