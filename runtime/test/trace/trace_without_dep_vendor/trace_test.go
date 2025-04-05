package trace_without_dep_vendor

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/example/hello/reverse"
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

func greet(s string) string {
	return "hello " + s
}
func TestGreet(t *testing.T) {
	result := greet(reverse.String("world"))
	if result != "hello dlrow" {
		t.Fatalf("result: %s", result)
	}
}
