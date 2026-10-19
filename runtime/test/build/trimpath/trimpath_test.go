package trimpath

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/functab"
)

var value = 1

func example() int {
	return value
}

func TestRecordedFile(t *testing.T) {
	const want = "example.com/trimpath/trimpath_test.go"
	fn := functab.InfoFunc(example)
	if fn == nil {
		t.Fatal("example not registered")
	}
	if fn.File != want {
		t.Fatalf("function file = %q, want %q", fn.File, want)
	}
	v := functab.InfoVar(&value)
	if v == nil {
		t.Fatal("value not registered")
	}
	if v.File != want {
		t.Fatalf("variable file = %q, want %q", v.File, want)
	}
}
