package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	_, err = xgoBuild([]string{"-o", tmpFile, "./testdata/hello_world"})
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

func getTempFile(pattern string) (string, error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", err
	}

	return filepath.Join(tmpDir, pattern), nil
}

func xgoBuild(args []string) (string, error) {
	buildArgs := append([]string{
		"run", "../cmd/xgo",
		"build",
		"--xgo-src",
		"../",
		"--sync-with-link",
	}, args...)
	cmd := exec.Command("go", buildArgs...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}
