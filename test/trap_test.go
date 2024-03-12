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

	origOut, err := xgoBuild([]string{"--no-instrument", "--project-dir", tmpDir, "./"}, &options{
		run:    true,
		noTrim: true,
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	origExpect := "A\nB\n"

	if origOut != origExpect {
		t.Fatalf("expect original output: %q, actual: %q", origExpect, origOut)
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
		fatalExecErr(t, err)
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

func fatalExecErr(t *testing.T, err error) {
	if err, ok := err.(*exec.ExitError); ok {
		t.Fatalf("%v", string(err.Stderr))
	}
	t.Fatalf("%v", err)
}
