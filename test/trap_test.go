package test

import (
	"os"
	"os/exec"
	"testing"
)

// go test -run TestTrap$ -v ./test
func TestTrap(t *testing.T) {
	origExpect := "A\nB\n"
	expectOut := "trap A\nA\nabort B\n"
	testTrap(t, "./testdata/trap", origExpect, expectOut)
}

func testTrap(t *testing.T, testDir string, origExpect string, expectOut string) {
	debug := false
	// debug := true
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpFile)

	if debug {
		tmpFile = "../trap.bin"
	}
	rootDir, tmpDir, err := tmpMergeRuntimeAndTest(testDir)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !debug {
		defer os.RemoveAll(rootDir)
	}

	// NOTE: the --no-instrument version should not use
	// the same cache as instrumented version, cache
	// cannot tell whether --no-instrument is applied
	origOut, err := xgoBuild([]string{"--no-instrument", "--project-dir", tmpDir, "./"}, &options{
		run:    true,
		noTrim: true,
		env: []string{
			"XGO_TEST_NO_INSTRUMENT=true",
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if origOut != origExpect {
		t.Fatalf("expect original output: %q, actual: %q", origExpect, origOut)
	}
	_, err = xgoBuild([]string{
		"-o", tmpFile,
		"--project-dir", tmpDir,
		// "-a", // debug
		// "--debug", "github.com/xhd2015/xgo/runtime/core/trap", "--vscode", "../.vscode", // debug
		"--",
		// "-gcflags=all=-N -l", // debug
		".",
	}, nil)
	if err != nil {
		fatalExecErr(t, err)
	}
	out, err := exec.Command(tmpFile).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			t.Fatalf("%v", string(err.Stderr))
		}
		t.Fatalf("%v", err)
	}
	outStr := string(out)
	// t.Logf("%s", outStr)

	if outStr != expectOut {
		t.Fatalf("expect output: %q, actual: %q", expectOut, outStr)
	}
}
