package revision

import (
	"fmt"
	"testing"
)

func TestIncrementNumber(t *testing.T) {
	newContent, err := IncrementNumber(`const NUMBER = 10`)
	if err != nil {
		t.Fatal(err)
	}
	expect := "const NUMBER = 11"
	if newContent != expect {
		t.Fatalf("expect new:%q, actual:%q", expect, newContent)
	}
}

func TestReplaceRevision(t *testing.T) {
	revision := "XYZ123A"
	newContent, err := ReplaceRevision(`const REVISION = "A123XYZ"`, revision)
	if err != nil {
		t.Fatal(err)
	}
	expect := fmt.Sprintf("const REVISION = %q", revision)
	if newContent != expect {
		t.Fatalf("expect new:%q, actual:%q", expect, newContent)
	}
}
