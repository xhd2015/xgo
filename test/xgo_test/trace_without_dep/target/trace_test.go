package target

import "testing"

func greet(s string) string {
	return "hello " + s
}
func TestGreet(t *testing.T) {
	result := greet("world")
	if result != "hello world" {
		t.Fatalf("result: %s", result)
	}
}
