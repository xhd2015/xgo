package edit

import "testing"

func TestBufferBytes(t *testing.T) {
	b := NewBuffer([]byte("func f() { return x }"))
	b.Replace(18, 19, "x_get()")
	b.Insert(10, " trap();")
	b.Delete(0, 0)
	b.Insert(18, "/*a*/")
	b.Insert(0, "// gen\n")
	got := b.Bytes()
	want := "// gen\nfunc f() { trap(); return /*a*/x_get() }"
	if string(got) != want {
		t.Fatalf("Bytes() = %q, want %q", got, want)
	}
	if cap(got) != len(want) {
		t.Errorf("cap = %d, want exactly %d", cap(got), len(want))
	}
	// a second call returns a new slice
	again := b.Bytes()
	if &again[0] == &got[0] || string(again) != want {
		t.Errorf("second Bytes() = %q, shares the first result: %v", again, &again[0] == &got[0])
	}

	shrink := NewBuffer([]byte("abcdef"))
	shrink.Delete(1, 5)
	if got := string(shrink.Bytes()); got != "af" {
		t.Errorf("delete: %q", got)
	}
}

func TestBufferBytesOverlap(t *testing.T) {
	b := NewBuffer([]byte("abcdef"))
	b.Replace(0, 4, "x")
	b.Replace(2, 6, "y")
	defer func() {
		if recover() == nil {
			t.Errorf("expected a panic for overlapping edits")
		}
	}()
	b.Bytes()
}
