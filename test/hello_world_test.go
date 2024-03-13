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
	output, err := buildAndRunOutput("./testdata/hello_world")
	if err != nil {
		t.Fatal(err)
	}
	expect := "hello world\n"
	if output != expect {
		t.Fatalf("expect output %q, actual:%q", expect, output)
	}
}
func buildAndRunOutput(program string) (output string, err error) {
	return buildAndRunOutputArgs([]string{program})
}

func buildAndRunOutputArgs(args []string) (output string, err error) {
	tmpFile, err := getTempFile("test")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpFile)
	buildArgs := []string{"-o", tmpFile}
	buildArgs = append(buildArgs, args...)
	_, err = xgoBuild(buildArgs, nil)
	if err != nil {
		return "", err
	}
	out, err := exec.Command(tmpFile).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
