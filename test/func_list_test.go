package test

import (
	"testing"
)

// go test -run TestFuncList -v ./test
func TestFuncList(t *testing.T) {
	goVersion, err := getGoVersion()
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	output, err := buildWithRuntimeAndOutput("./testdata/func_list", buildRuntimeOpts{})
	if err != nil {
		t.Fatal(getErrMsg(err))
	}

	// t.Logf("%s", output)

	expectLines := []string{
		"func:strconv FormatBool",
		"func:time Now",
		"func:os MkdirAll",
		"func:fmt Printf",
		"func:strings Split",
	}

	// go1.18, with generic
	if goVersion.Major >= 1 && goVersion.Minor >= 18 {
		expectLines = append(expectLines, []string{
			"func:main generic",
			"func:main List.size",
		}...)
	}
	expectLines = append(expectLines, []string{
		"func:main example",
		"func:main someInt.value",
		// "func:main (*someInt).inc", // this output is replaced with a simplier display name
		"func:main someInt.inc",
	}...)

	expectSequence(t, output, expectLines)
}
