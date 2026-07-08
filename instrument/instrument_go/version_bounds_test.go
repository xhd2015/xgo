package instrument_go

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/instrument/instrument_go/instrument_unifiedtest"
	"github.com/xhd2015/xgo/support/goinfo"
)

func TestInstrumentExecSupportsGo126(t *testing.T) {
	goroot := t.TempDir()
	execDir := filepath.Join(goroot, "src", "cmd", "go", "internal", "work")
	if err := os.MkdirAll(execDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(execDir, "exec.go"), []byte("package work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := instrumentExec(goroot, &goinfo.GoVersion{Major: 1, Minor: 26})
	if err != nil && strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("go1.26 must be supported by legacy instrumentExec, got: %v", err)
	}
}

func TestInstrumentExecRejectsGo127(t *testing.T) {
	err := instrumentExec(t.TempDir(), &goinfo.GoVersion{Major: 1, Minor: 27})
	if err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("go1.27 must be rejected (legacy capped at go1.26), got: %v", err)
	}
}

func TestUnifySupportsGo126(t *testing.T) {
	goroot := t.TempDir()
	testDir := filepath.Join(goroot, "src", "cmd", "go", "internal", "test")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "test.go"), []byte("package test\nfunc runTest() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loadDir := filepath.Join(goroot, "src", "cmd", "go", "internal", "load")
	if err := os.MkdirAll(loadDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loadDir, "test.go"), []byte("package load\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	jsonDir := filepath.Join(goroot, "src", "cmd", "internal", "test2json")
	if err := os.MkdirAll(jsonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jsonDir, "test2json.go"), []byte("package test2json\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := instrument_unifiedtest.Unify(goroot, &goinfo.GoVersion{Major: 1, Minor: 26})
	if err != nil && strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("go1.26 must be supported by legacy unified test instrument, got: %v", err)
	}
}

func TestGO_VERSION_26Constant(t *testing.T) {
	if goinfo.GO_VERSION_26 != 26 {
		t.Fatalf("GO_VERSION_26 = %d, want 26", goinfo.GO_VERSION_26)
	}
}