package test

import (
	"testing"
)

// go test -run TestFuncList -v ./test
func TestFuncList(t *testing.T) {
	t.Parallel()
	goVersion, err := getGoVersion()
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	output, err := buildWithRuntimeAndOutput("./testdata/func_list", buildRuntimeOpts{})
	if err != nil {
		t.Fatal(getErrMsg(err))
	}

	// t.Logf("%s", output)

	if false {
		// NOTE: since xgo v1.0.3, std libs
		// are not included in functab
		expectStdLines := []string{
			"func:strconv FormatBool",
			"func:time Now",
			"func:os MkdirAll",
			"func:fmt Printf",
			"func:strings Split",
		}
		expectContains(t, output, expectStdLines)
	}

	expectMainLines := []string{
		"func:main example",
		"func:main someInt.value",
		// "func:main (*someInt).inc", // this output is replaced with a simpler display name
		"func:main someInt.inc",
	}

	expectSequence(t, output, expectMainLines)

	// go1.18, with generic
	if goVersion.Major >= 1 && goVersion.Minor >= 18 {
		expectGenericLines := []string{
			"func:main generic",
			"func:main List.size",
		}
		expectSequence(t, output, expectGenericLines)
	}
}
