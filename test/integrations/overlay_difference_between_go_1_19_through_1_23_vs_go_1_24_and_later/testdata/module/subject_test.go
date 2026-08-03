package target

import "testing"

func TestValue(t *testing.T) {
	if got, want := Value(), "caller replacement"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}
