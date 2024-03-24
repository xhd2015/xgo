package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/osinfo"
)

// go test -run TestTrap$ -v ./test
func TestTrap(t *testing.T) {
	t.Parallel()
	origExpect := "A\nB\n"
	expectOut := "trap A\nA\nabort B\n"
	testTrap(t, "./testdata/trap", origExpect, expectOut)
}

// go test -run TestTrapNormalBuildShouldWarn -v ./test
func TestTrapNormalBuildShouldWarn(t *testing.T) {
	t.Parallel()
	expectOrigStderr := "failed to link __xgo_link_set_trap"

	var origStderr bytes.Buffer
	runAndCheckInstrumentOutput(t, "./testdata/trap", func(output string) error {
		stderr := origStderr.String()
		// t.Logf("orig stderr: %s", stderr)
		if !strings.Contains(stderr, expectOrigStderr) {
			return fmt.Errorf("expect orig stderr contains: %q, actual: %q", expectOrigStderr, stderr)
		}
		return nil
	}, func(output string) error {
		t.Fatalf("runInstrument set to false, should not be called")
		return nil
	}, testTrapOpts{
		expectOrigErr:       false,
		withNoInstrumentEnv: false,
		runInstrument:       false,
		origStderr:          &origStderr,
	})
}

func testTrap(t *testing.T, testDir string, origExpect string, expectOut string) {
	runAndCompareInstrumentOutput(t, testDir, origExpect, expectOut, testTrapOpts{
		expectOrigErr:       false,
		withNoInstrumentEnv: true,
		runInstrument:       true,
	})
}

func testTrapWithTest(t *testing.T, testDir string, orig func(output string) error, instr func(output string) error) {
	runAndCheckInstrumentOutput(t, testDir, orig, instr, testTrapOpts{
		test:                true,
		expectOrigErr:       false,
		withNoInstrumentEnv: true,
		runInstrument:       true,
	})
}

type testTrapOpts struct {
	test                bool
	expectOrigErr       bool
	withNoInstrumentEnv bool
	runInstrument       bool

	origStderr io.Writer
}

func runAndCompareInstrumentOutput(t *testing.T, testDir string, origExpect string, expectOut string, opts testTrapOpts) {
	runAndCheckInstrumentOutput(t, testDir, func(output string) error {
		if opts.expectOrigErr {
			if expectOut == "" || !strings.Contains(output, expectOut) {
				t.Fatalf("expect build err contains: %q, actual: %s", expectOut, output)
			}
			return nil
		}

		if output != origExpect {
			t.Fatalf("expect original output: %q, actual: %q", origExpect, output)
		}
		return nil
	}, func(output string) error {
		if output != expectOut {
			t.Fatalf("expect output: %q, actual: %q", expectOut, output)
		}
		return nil
	}, opts)
}
func runAndCheckInstrumentOutput(t *testing.T, testDir string, orig func(output string) error, instr func(output string) error, opts testTrapOpts) {
	debug := false
	// debug := true
	tmpFile, err := getTempFile("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exeSuffix := osinfo.EXE_SUFFIX
	tmpFile += exeSuffix
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
		env = append(env, "XGO_TEST_HAS_INSTRUMENT=false")
	}
	var testArgs []string
	origCmd := xgoCmd_run
	if opts.test {
		origCmd = xgoCmd_test
		testArgs = append(testArgs, "-v")
	}
	origOut, err := runXgo(append([]string{"--no-instrument", "--project-dir", tmpDir, "./"}, testArgs...), &options{
		xgoCmd:       origCmd,
		noTrim:       true,
		env:          env,
		noPipeStderr: opts.expectOrigErr,
		stderr:       opts.origStderr,
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

		if err := orig(extStdErr); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := orig(origOut); err != nil {
		t.Fatal(err)
	}

	if !opts.runInstrument {
		return
	}

	instrCmd := xgoCmd_build
	var instrArgs []string
	if opts.test {
		instrCmd = xgoCmd_testBuild
		instrArgs = append(instrArgs, "-test.v")
	}
	_, err = runXgo([]string{
		"-o", tmpFile,
		"--project-dir", tmpDir,
		// "-a", // debug
		// "--debug", "github.com/xhd2015/xgo/runtime/core/trap", "--vscode", "../.vscode", // debug
		"--",
		// "-gcflags=all=-N -l", // debug
		".",
	}, &options{
		xgoCmd: instrCmd,
	})
	if err != nil {
		fatalExecErr(t, err)
	}
	out, err := exec.Command(tmpFile, instrArgs...).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			t.Fatalf("%v", string(err.Stderr))
		}
		t.Fatalf("%v", err)
	}
	outStr := string(out)
	// t.Logf("%s", outStr)

	if err := instr(outStr); err != nil {
		t.Fatal(err)
	}
}
