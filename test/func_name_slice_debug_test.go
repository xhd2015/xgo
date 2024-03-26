// this test run xgo_test/func_name_slice with
// -gcflags=all=-N -l.
// the purepose is to ensure when in dubugging mode,
// operation on the function name returned by runtime
// does not cause panic.
//
// Rational: previously there  were problem slicing
// the function name in debugging mode

package test

import "testing"

// go test -run TestFuncNameSliceShouldWorkWithDebug -v ./test
func TestFuncNameSliceShouldWorkWithDebug(t *testing.T) {
	t.Parallel()
	output, err := buildAndRunOutputArgs([]string{"./xgo_test/func_name_slice"}, buildAndOutputOptions{
		buildTest: true,
		buildArgs: []string{"-gcflags=all=-N -l"},
	})
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	// t.Logf("output: %s", output)
	expect := "PASS\n"
	if output != expect {
		t.Fatalf("expect output %q, actual:%q", expect, output)
	}
}
