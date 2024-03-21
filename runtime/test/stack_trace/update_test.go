package stack_trace

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Enable()
}

// xgo test -run TestUpdateUserInfo -v ./test/stack_trace
func TestUpdateUserInfo(t *testing.T) {
	oldUserName := "old user"
	actualName, err := UpdateUserInfo(oldUserName)
	if err != nil {
		t.Fatal(err)
	}
	if actualName == oldUserName {
		t.Fatalf("expect user name: %s, actual: %s", oldUserName, actualName)
	}
	newActualName, err := UpdateUserInfo(actualName)
	if err != nil {
		t.Fatal(err)
	}
	if newActualName != actualName {
		t.Fatalf("expect new user name: %s, actual: %s", actualName, newActualName)
	}
}

// xgo test -run TestDeleteUserInfoPanic -v ./test/stack_trace
func TestDeleteUserInfoPanic(t *testing.T) {
	DeleteUserInfo("xhd2015")
}
