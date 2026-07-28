package instrument_go

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestInstrumentExecRejectsGo127(t *testing.T) {
	err := instrumentExec(t.TempDir(), &goinfo.GoVersion{Major: 1, Minor: 27})
	if err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("go1.27 must be rejected (legacy capped at go1.26), got: %v", err)
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