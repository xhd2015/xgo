package main

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trap/flags"
)

// go run -tags dev ./cmd/xgo test -v --project-dir runtime/test/trap/flags/persistent_after_build/testdata --strace --strace-dir /tmp ./

// separate:
//
//	go run -tags dev ./cmd/xgo test -c --project-dir runtime/test/trap/flags/persistent_after_build/testdata --strace --strace-dir /tmp -o test.bin
//
// ./test.bin -test.v
func TestFlags(t *testing.T) {
	if flags.STRACE != "on" {
		t.Errorf("STRACE expect: %s, actual: %s", "on", flags.STRACE)
	}
	if flags.STRACE_DIR != "/tmp" {
		t.Errorf("STRACE_DIR expect: %s, actual: %s", "/tmp", flags.STRACE_DIR)
	}
}
