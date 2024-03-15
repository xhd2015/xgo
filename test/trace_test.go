package test

import (
	"os/exec"
	"testing"
)

// go test -run TestTrace -v ./test
func TestTrace(t *testing.T) {
	output, err := buildWithRuntimeAndOutput("./testdata/trace", buildRuntimeOpts{
		runEnv: []string{
			"XGO_TRACE_DIR=stdout",
		},
	})
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			t.Logf("stderr: %s", string(err.Stderr))
		}
		t.Fatal(err)
	}

	// t.Logf("%s", output)
	expectLines := []string{
		// output
		"A\nB\nC\nC\n",

		// trace
		"FuncInfo",
		"main.main",
		"FuncInfo",
		"main.A",
		"FuncInfo",
		"main.B",
		"FuncInfo",
		"main.C",
		"FuncInfo",
		"main.C",
	}
	expectSequence(t, output, expectLines)
}
