package goinfo

import (
	"reflect"
	"testing"
)

func TestTestPackages(t *testing.T) {
	// `go list -deps -test ./p` where p_test imports q, and q imports p
	pkgs := []*Package{
		{ImportPath: "m/dep", DepOnly: true},
		{ImportPath: "m/p", Match: []string{"./p"}},
		{ImportPath: "m/p [m/p.test]", ForTest: "m/p", Match: []string{"./p"}},
		{ImportPath: "m/testutil", DepOnly: true},
		{ImportPath: "m/q [m/p.test]", ForTest: "m/p", DepOnly: true},
		{ImportPath: "m/p_test [m/p.test]", ForTest: "m/p"},
		{ImportPath: "m/p.test", Name: "main"},
	}
	var got []string
	for _, pkg := range testPackages(pkgs) {
		if pkg.ForTest != "" {
			t.Errorf("%s: ForTest not cleared", pkg.ImportPath)
		}
		got = append(got, pkg.ImportPath)
	}
	want := []string{"m/dep", "m/p", "m/testutil", "m/q"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("testPackages = %v, want %v", got, want)
	}
}
