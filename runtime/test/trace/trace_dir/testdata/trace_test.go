package testdata

import "testing"

func TestGreet(t *testing.T) {
	v := greet("world")
	if v != "hello world" {
		t.Errorf("greet: %s", v)
	}
}

func greet(s string) string {
	return "hello " + s
}
