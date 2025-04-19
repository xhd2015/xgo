package atomic_generic

import "testing"

func TestAtomicGeneric(t *testing.T) {
	if testAtomicPointer != nil {
		res1 := testAtomicPointer()
		if res1 != 10 {
			t.Fatalf("expect res1 to be 10, actual: %d", res1)
		}
	}
	res2 := testLocalPointer()
	if res2 != 10 {
		t.Fatalf("expect res2 to be 10, actual: %d", res2)
	}
}
