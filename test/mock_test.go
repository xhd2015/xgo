package test

import (
	"fmt"
	"os/exec"
	"testing"
)

// go test -run TestMockArg -v ./test
func TestMockArg(t *testing.T) {
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

// go test -run TestMockResult -v ./test
func TestMockResult(t *testing.T) {
	t.Parallel()
	expectOrig := "before mock: add(5,2)=7\nafter mock: add(5,2)=7\n"
	expectInstrument := "before mock: add(5,2)=7\nafter mock: add(5,2)=3\n"
	err := testNoInstrumentAndInstrumentOutput("./testdata/mock_res", expectOrig, expectInstrument)
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
