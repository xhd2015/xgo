package trace_without_dep

import (
	"testing"
)

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
