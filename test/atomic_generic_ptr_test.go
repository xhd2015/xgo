//go:build go1.18
// +build go1.18

package test

import "testing"

// this test is specifically for go1.19, it has problems with generic
// function trap when instantiated from other packages

// go test -run TestAtomicGenericPtr -v ./test
func TestAtomicGenericPtr(t *testing.T) {
	t.Parallel()
	goVersion, err := getGoVersion()
	if err != nil {
		t.Fatal(getErrMsg(err))
	}
	output, err := buildAndRunOutput("./testdata/atomic_generic")
	if err != nil {
		t.Fatal(err)
	}
	expect := "local load: 10\n"
	if goVersion.Major == 1 && goVersion.Minor >= 19 {
		expect = "atomic sload: 10\n" + expect
	}
	if output != expect {
		t.Fatalf("expect output %q, actual:%q", expect, output)
	}
}
