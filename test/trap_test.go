package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// go test -run TestTrap$ -v ./test
func TestTrap(t *testing.T) {
	origExpect := "A\nB\n"
	expectOut := "trap A\nA\nabort B\n"
	testTrap(t, "./testdata/trap", origExpect, expectOut)
}

// go test -run TestTrapNormalBuildShouldFail -v ./test
func TestTrapNormalBuildShouldFail(t *testing.T) {
	expectOut := "panic: failed to link __xgo_link_set_trap"
	testTrapWithOpts(t, "./testdata/trap", "", expectOut, testTrapOpts{
		expectOrigErr:       true,
		withNoInstrumentEnv: false,
		runInstrument:       false,
	})
}

func testTrap(t *testing.T, testDir string, origExpect string, expectOut string) {
	testTrapWithOpts(t, testDir, origExpect, expectOut, testTrapOpts{
		expectOrigErr:       false,
		withNoInstrumentEnv: true,
		runInstrument:       true,
	})
}

type testTrapOpts struct {
	expectOrigErr       bool
	withNoInstrumentEnv bool
	runInstrument       bool
}

func testTrapWithOpts(t *testing.T, testDir string, origExpect string, expectOut string, opts testTrapOpts) {
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
	var env []string
	if opts.withNoInstrumentEnv {
		env = append(env, "XGO_TEST_NO_INSTRUMENT=true")
	}
	origOut, err := runXgo([]string{"--no-instrument", "--project-dir", tmpDir, "./"}, &options{
		run:          true,
		noTrim:       true,
		env:          env,
		noPipeStderr: opts.expectOrigErr,
	})
	if opts.expectOrigErr {
		if err == nil {
			t.Fatalf("expect build no instrument err, actual no err")
		}
		// errOut
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("expect build err be *exec.ExitError, actual: %T %v", err, err)
		}
		extStdErr := string(exitErr.Stderr)
		if expectOut == "" || !strings.Contains(extStdErr, expectOut) {
			t.Fatalf("expect build err contains: %q, actual: %s", expectOut, extStdErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("%v", err)
	}

	if origOut != origExpect {
		t.Fatalf("expect original output: %q, actual: %q", origExpect, origOut)
	}
	if !opts.runInstrument {
		return
	}
	_, err = runXgo([]string{
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
