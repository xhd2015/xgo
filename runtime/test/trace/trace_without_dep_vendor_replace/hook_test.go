//go:build ignore
// +build ignore

package trace_without_dep_vendor_replace

import (
	"os"
	"testing"
)

const traceFile = "TestGreet.json"

func TestPreCheck(t *testing.T) {
	err := os.RemoveAll(traceFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostCheck(t *testing.T) {
	_, err := os.Stat(traceFile)
	if err != nil {
		t.Fatalf("expect %s, actual: %v", traceFile, err)
	}
}
