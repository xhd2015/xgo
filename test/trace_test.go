package test

import (
	"os/exec"
	"testing"
)

// go test -run TestTrace -v ./test
func TestTrace(t *testing.T) {
	t.Parallel()
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
		`"Pkg":"main"`,
		`"IdentityName":"main"`,
		"FuncInfo",
		`"Pkg":"main"`,
		`"IdentityName":"A"`,
		"FuncInfo",
		`"Pkg":"main"`,
		`"IdentityName":"B"`,
		"FuncInfo",
		`"Pkg":"main"`,
		`"IdentityName":"C"`,
		"FuncInfo",
		`"Pkg":"main"`,
		`"IdentityName":"C"`,
	}
	expectSequence(t, output, expectLines)
}
