package stack_trace

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Enable()
}

// xgo test ./test/stack_trace
func TestUpdateUseInfo(t *testing.T) {
	oldUserName := "old user"
	actualName, err := UpdateUseInfo(oldUserName)
	if err != nil {
		t.Fatal(err)
	}
	if actualName == oldUserName {
		t.Fatalf("expect user name: %s, actual: %s", oldUserName, actualName)
	}
	newActualName, err := UpdateUseInfo(actualName)
	if err != nil {
		t.Fatal(err)
	}
	if newActualName != actualName {
		t.Fatalf("expect new user name: %s, actual: %s", actualName, newActualName)
	}
}
