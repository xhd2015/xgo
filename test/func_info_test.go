package test

import "testing"

// go test -run TestFuncInfo -v ./test
func TestFuncInfo(t *testing.T) {
	output, err := buildWithRuntimeAndOutput("./testdata/func_info", buildRuntimeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("%s", output)

	expect := "fullName: main.example\nargs: [a]\n"
	if output != expect {
		t.Logf("expect output %q, actual: %q", expect, output)
	}
}
