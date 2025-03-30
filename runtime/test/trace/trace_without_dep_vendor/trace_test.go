package trace_without_dep_vendor

import (
	"testing"

	"golang.org/x/example/hello/reverse"
)

func greet(s string) string {
	return "hello " + s
}
func TestGreet(t *testing.T) {
	result := greet(reverse.String("world"))
	if result != "hello dlrow" {
		t.Fatalf("result: %s", result)
	}
}
