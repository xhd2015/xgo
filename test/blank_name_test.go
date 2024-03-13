package test

import (
	"os"
	"os/exec"
	"testing"
)

// go test -run TestBlankName -v ./test
func TestBlankName(t *testing.T) {
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpFile)
	_, err = xgoBuild([]string{"-o", tmpFile,
		// "-a", // debug
		"./testdata/blank_name"}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	out, err := exec.Command(tmpFile).Output()
	if err != nil {
		t.Fatalf("%v", err)
	}
	outStr := string(out)
	expect := "test blank name\n"
	if outStr != expect {
		t.Fatalf("expect output %q, actual:%q", expect, outStr)
	}
}
