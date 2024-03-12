package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// go test -run TestFuncList -v ./test
func TestFuncList(t *testing.T) {
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpFile)

	tmpDir, funcListDir, err := tmpMergeRuntimeAndTest("./testdata/func_list")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = xgoBuild([]string{
		"-o", tmpFile,
		"--project-dir", funcListDir,
		".",
	}, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	out, err := exec.Command(tmpFile).Output()
	if err != nil {
		t.Fatalf("%v", err)
	}
	outStr := string(out)
	// t.Logf("%s", outStr)

	expectLines := []string{
		"func:strconv FormatBool",
		"func:time Now",
		"func:os MkdirAll",
		"func:fmt Printf",
		"func:strings Split",
	}
	lines := strings.Split(outStr, "\n")
	for _, expectLine := range expectLines {
		if !containsLine(lines, expectLine) {
			t.Fatalf("expect %s not found", expectLine)
		}
	}
}

func containsLine(lines []string, line string) bool {
	for _, t := range lines {
		if t == line {
			return true
		}
	}
	return false
}
