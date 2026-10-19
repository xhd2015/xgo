package instrument_go

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestInstrumentExecRejectsGo127AndLater(t *testing.T) {
	for _, minor := range []int{27, 28} {
		err := instrumentExec(t.TempDir(), &goinfo.GoVersion{Major: 1, Minor: minor})
		if err == nil || !strings.Contains(err.Error(), "unsupported version") {
			t.Fatalf("go1.%d must be rejected (legacy capped at go1.26), got: %v", minor, err)
		}
	}
}

func TestGO_VERSION_26Constant(t *testing.T) {
	if goinfo.GO_VERSION_26 != 26 {
		t.Fatalf("GO_VERSION_26 = %d, want 26", goinfo.GO_VERSION_26)
	}
}

func TestGO_VERSION_27Constant(t *testing.T) {
	if goinfo.GO_VERSION_27 != 27 {
		t.Fatalf("GO_VERSION_27 = %d, want 27", goinfo.GO_VERSION_27)
	}
}
