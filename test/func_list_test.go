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

	t.Logf("%s", output)

	expectStdLines := []string{
		"func:strconv FormatBool",
		"func:time Now",
		"func:os MkdirAll",
		"func:fmt Printf",
		"func:strings Split",
	}
	expectContains(t, output, expectStdLines)

	var expectMainLines []string
	// go1.18, with generic
	if goVersion.Major >= 1 && goVersion.Minor >= 18 {
		expectMainLines = append(expectMainLines, []string{
			"func:main generic",
			"func:main List.size",
		}...)
	}
	expectMainLines = append(expectMainLines, []string{
		"func:main example",
		"func:main someInt.value",
		// "func:main (*someInt).inc", // this output is replaced with a simplier display name
		"func:main someInt.inc",
	}...)

	expectSequence(t, output, expectMainLines)
}
