package target

import "testing"

func TestCallerReplacementWasCompiled(t *testing.T) {
	if got := Value(); got != "caller replacement" {
		t.Fatalf("Value() = %q, want caller replacement", got)
	}
}
