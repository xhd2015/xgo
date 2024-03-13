package test

import (
	"testing"
)

// the hello world test verify if
// the xgo can compile a source file
// correctly
// it serves as a smoke test.

// go test -run TestHelloWorld -v ./test
func TestHelloWorld(t *testing.T) {
	output, err := buildAndRunOutput("./testdata/hello_world")
	if err != nil {
		t.Fatal(err)
	}
	expect := "hello world\n"
	if output != expect {
		t.Fatalf("expect output %q, actual:%q", expect, output)
	}
}
