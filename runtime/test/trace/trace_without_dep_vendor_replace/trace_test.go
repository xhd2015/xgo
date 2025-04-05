package trace_without_dep_vendor_replace

import (
	"fmt"
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/trace_without_dep_vendor_replace/lib"
)

const traceFile = "TestGreet.json"

func TestMain(m *testing.M) {
	err := os.RemoveAll(traceFile)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	exitCode := m.Run()
	_, err = os.Stat(traceFile)
	if err != nil {
		fmt.Printf("expect %s, actual: %v", traceFile, err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func TestGreet(t *testing.T) {
	result := lib.Greet("world")
	if result != "hello world" {
		t.Fatalf("result: %s", result)
	}
}
