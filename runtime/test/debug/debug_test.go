// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestDebug(t *testing.T) {
	mock.Patch(greet, func(s string) string {
		return "mock " + s
	})
	greet("world")
}

func greet(s string) string {
	return "hello " + s
}
