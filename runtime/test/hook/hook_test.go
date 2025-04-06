package hook

import (
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/hook"
	"github.com/xhd2015/xgo/runtime/test/hook/id"
	"github.com/xhd2015/xgo/runtime/test/hook/pkg"
)

var testMainID int

func TestMain(m *testing.M) {
	testMainID = id.Next()
	os.Exit(m.Run())
}

var testInitID int
var testInsideID int

func init() {
	hook.OnInitFinished(func() {
		testInitID = id.Next()
	})
}

func TestOnInitFinished(t *testing.T) {
	hook.OnInitFinished(func() {
		testInsideID = id.Next()
	})

	if !hook.InitFinished() {
		t.Error("init not finished inside test")
	}

	if pkg.ID != 1 {
		t.Errorf("pkg.ID = %d, want 1", pkg.ID)
	}
	if mainInitID != 2 {
		t.Errorf("initID = %d, want 2", mainInitID)
	}
	if testInitID != 3 {
		t.Errorf("testInitID = %d, want 3", testInitID)
	}
	if testMainID != 4 {
		t.Errorf("testMainID = %d, want 4", testMainID)
	}
	if testInsideID != 0 {
		t.Errorf("testInsideID = %d, want 0", testInsideID)
	}
}
