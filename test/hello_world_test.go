package test

import (
	"os"
	"os/exec"
	"testing"
)

// the hello world test verify if
// the xgo can compile a source file
// correctly
// it serves as a smoke test.

// go test -run TestHelloWorld -v ./test
func TestHelloWorld(t *testing.T) {
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpFile)
	_, err = xgoBuild([]string{"-o", tmpFile, "./testdata/hello_world"}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	out, err := exec.Command(tmpFile).Output()
	if err != nil {
		t.Fatalf("%v", err)
	}
	outStr := string(out)
	expect := "hello world\n"
	if outStr != expect {
		t.Fatalf("expect output %q, actual:%q", expect, outStr)
	}
}
