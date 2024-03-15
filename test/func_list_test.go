package test

import (
	"testing"
)

// go test -run TestFuncList -v ./test
func TestFuncList(t *testing.T) {
	output, err := buildWithRuntimeAndOutput("./testdata/func_list", buildRuntimeOpts{})
	if err != nil {
		t.Fatal(err)
	}

	// t.Logf("%s", output)

	expectLines := []string{
		"func:strconv FormatBool",
		"func:time Now",
		"func:os MkdirAll",
		"func:fmt Printf",
		"func:strings Split",
		"func:main example",
		"func:main someInt.value",
		// "func:main (*someInt).inc", // this output is replaced with a simplier display name
		"func:main someInt.inc",
	}
	expectSequence(t, output, expectLines)
}
