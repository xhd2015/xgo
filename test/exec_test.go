package test

import (
	"testing"

	"github.com/xhd2015/xgo/support/cmd"
)

// go test -run TestExec -v ./test
func TestExec(t *testing.T) {
	t.Parallel()
	goVersion, err := cmd.Output("go", "version")
	if err != nil {
		t.Fatal(err)
	}

	goVersionExec, err := xgoExec("go", "version")
	if err != nil {
		t.Fatal(err)
	}

	if goVersion != goVersionExec {
		t.Fatalf("expect exec go version: %q, actual: %q", goVersion, goVersionExec)
	}
}
