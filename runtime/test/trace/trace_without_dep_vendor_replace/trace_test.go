package trace_without_dep_vendor_replace

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/test/trace_without_dep_vendor_replace/lib"
)

func TestGreet(t *testing.T) {
	result := lib.Greet("world")
	if result != "hello world" {
		t.Fatalf("result: %s", result)
	}
}
