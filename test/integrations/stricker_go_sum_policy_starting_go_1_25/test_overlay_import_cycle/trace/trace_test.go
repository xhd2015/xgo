package trace

import "testing"

func TestTrace(t *testing.T) {
	c := Config{FilterTrace: nil}
	if c.FilterTrace != nil {
		t.Error("expected nil FilterTrace")
	}
}
