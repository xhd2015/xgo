package test

import (
	"os"
	"os/exec"
	"testing"
)

// go test -run TestTrap -v ./test
func TestTrap(t *testing.T) {
	debug := false
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.RemoveAll(tmpFile)

	if debug {
		tmpFile = "../trap.bin"
	}
	rootDir, tmpDir, err := tmpMergeRuntimeAndTest("./testdata/trap")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !debug {
		defer os.RemoveAll(rootDir)
	}

	_, err = xgoBuild([]string{
		"-o", tmpFile,
		"--project-dir", tmpDir,
		// "-a", // debug
		// "--debug", "main", "--vscode", "../.vscode", // debug
		"--",
		// "-gcflags=all=-N -l", // debug
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

	expectOut := "trap A\nA\nabort B\n"
	if outStr != expectOut {
		t.Fatalf("expect output: %q, actual: %q", expectOut, outStr)
	}
}
