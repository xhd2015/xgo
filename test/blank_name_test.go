package test

import (
	"testing"
)

// go test -run TestBlankName -v ./test
func TestBlankName(t *testing.T) {
	t.Parallel()
	output, err := buildAndRunOutput("./testdata/blank_name")
	if err != nil {
		t.Fatal(err)
	}
	expect := "test blank name\n"
	if output != expect {
		t.Fatalf("expect output %q, actual:%q", expect, output)
	}
}
