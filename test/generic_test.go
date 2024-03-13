package test

import (
	"testing"
)

// go test -run TestGeneric -v ./test
func TestGeneric(t *testing.T) {
	output, err := buildAndRunOutputArgs([]string{"--project-dir", "./testdata/generic_param", "./"}, buildAndOutputOptions{})
	if err != nil {
		t.Fatal(err)
	}
	expect := "[5 4 3 2 1]\n"
	if output != expect {
		t.Fatalf("expect output %q, actual:%q", expect, output)
	}
}
