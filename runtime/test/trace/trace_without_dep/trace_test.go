package trace_without_dep

import (
	"fmt"
	"os"
	"testing"
)

const traceFile = "TestGreet.json"

// --strace flag
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

// run this test, and check
func greet(s string) string {
	return "hello " + s
}

func TestGreet(t *testing.T) {
	result := greet("world")
	if result != "hello world" {
		t.Fatalf("result: %s", result)
	}
}
