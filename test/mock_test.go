package test

import (
	"fmt"
	"os/exec"
	"testing"
)

// go test -run TestMockInfo -v ./test
func TestMockInfo(t *testing.T) {
	t.Parallel()
	expectOrig := "hello world\n"
	expectInstrument := "hello mock:world\n"
	err := testNoInstrumentAndInstrumentOutput("./testdata/mock", expectOrig, expectInstrument)
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			t.Logf("stderr: %s", string(err.Stderr))
		}
		t.Fatal(err)
	}
}
func testNoInstrumentAndInstrumentOutput(dir string, expectOrig string, expectInstrument string) error {
	origOutput, err := buildWithRuntimeAndOutput(dir, buildRuntimeOpts{
		xgoBuildArgs: []string{"--no-instrument"},
		runEnv: []string{
			"XGO_TEST_NO_INSTRUMENT=true",
		},
	})
	if err != nil {
		return err
	}
	// t.Logf("%s", output)

	if origOutput != expectOrig {
		return fmt.Errorf("expect original output %q, actual: %q", expectOrig, origOutput)
	}

	instrumentOutput, err := buildWithRuntimeAndOutput(dir, buildRuntimeOpts{})
	if err != nil {
		return err
	}

	if instrumentOutput != expectInstrument {
		return fmt.Errorf("expect instrument output %q, actual: %q", expectInstrument, instrumentOutput)
	}
	return nil
}
