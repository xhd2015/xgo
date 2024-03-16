package test

import (
	"strings"
	"testing"
)

// go test -run TestFuncInfo -v ./test
func TestFuncInfo(t *testing.T) {
	goVersion, err := getGoVersion()
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	output, err := buildWithRuntimeAndOutput("./testdata/func_info", buildRuntimeOpts{
		// debug: true,
	})
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	// t.Logf("%s", output)

	expectNonGeneric := "example fullName: main.example\nexample args: [a]\n"
	if !strings.HasPrefix(output, expectNonGeneric) {
		t.Fatalf("expect output prefix %q, actual: %q", expectNonGeneric, output)
	}

	output = output[len(expectNonGeneric):]

	// go1.18, with generic
	if goVersion.Major >= 1 && goVersion.Minor >= 18 {
		expectGeneric := "generic func info\n(*List).Size fullName: \n(*List).Size args: []\nHello fullName: \nHello args: [v]\n"
		if output != expectGeneric {
			t.Fatalf("expect output generic %q, actual: %q", expectGeneric, output)
		}
	} else {
		if output != "" {
			t.Fatalf("expect output no more: %q", output)
		}
	}
}
