package test

import (
	"testing"
)

// go test -run TestTraceJSONOutput -v ./test
func TestTraceJSONOutput(t *testing.T) {
	t.Parallel()
	output, err := buildWithRuntimeAndOutput("./testdata/trace", buildRuntimeOpts{
		runEnv: []string{
			"XGO_TRACE_OUTPUT=stdout",
		},
	})
	if err != nil {
		t.Fatal(getErrMsg(err))
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

// go test -run TestTracePanicCapture -v ./test
func TestTracePanicCapture(t *testing.T) {
	t.Parallel()
	output, err := buildWithRuntimeAndOutput("./testdata/trace_panic_peek", buildRuntimeOpts{
		runEnv: []string{
			"XGO_TRACE_OUTPUT=stdout",
		},
	})
	if err != nil {
		t.Fatal(getErrMsg(err))
	}

	// t.Logf("%s", output)

	// output
	expectOutputLines := []string{
		"call: main\n",
		"call: Work\n",
		"call: doWork\n",
		"main panic: Work panic: doWork panic",
	}
	expectSequence(t, output, expectOutputLines)

	expectTraceOutput := []string{
		// main
		`{"Begin":"`,
		`"Name":"main"`,
		`"Panic":false,`,

		// Work
		`"Name":"Work",`,
		`"Panic":true,`,
		`"Error":"Work panic: doWork panic",`,

		// doWork
		`"Name":"doWork",`,
		`"Panic":true,`,
		`"Error":"panic: doWork panic"`,
	}
	expectSequence(t, output, expectTraceOutput)
}
