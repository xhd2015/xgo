package mock_closuer

import (
	"fmt"
	"testing"
)

func TestImmediatelyCalledClosure(t *testing.T) {
	// assert the compiler can compile
	// and run this code
	text := func(a int) string {
		return fmt.Sprintf("%d", a)
	}(10)
	if text != "10" {
		t.Logf("expect text to be %s, actual: %s", "10", text)
	}
}

func TestDeferNoArgCall(t *testing.T) {
	var b int
	defer func() {
		if b != 10 {
			t.Fatalf("expect b to be %d, actual: %d", 10, b)
		}
	}()

	b = 10
}

func TestDeferArgCall(t *testing.T) {
	var b int
	defer func(b int) {
		if b != 10 {
			t.Fatalf("expect b to be %d, actual: %d", 10, b)
		}
	}(10)

	b = 11
	_ = b
}
