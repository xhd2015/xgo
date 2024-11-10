package mock_simple

import (
	"runtime"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func greet(s string) string {
	return "hello " + s
}

func TestMockSimple(t *testing.T) {
	var skipped bool
	func() {
		defer func() {
			if e := recover(); e != nil {
				if runtime.GOOS == "windows" {
					t.Skipf("skip TestMockSimple on windows")
					skipped = true
				}
			}
		}()
		mock.Patch(greet, func(s string) string {
			return "mock " + s
		})
	}()
	if skipped {
		return
	}
	s := greet("world")
	if s != "mock world" {
		t.Fatalf("expect greet(): %q, actual: %q", "mock world", s)
	}
}
