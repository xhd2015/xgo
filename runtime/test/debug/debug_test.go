// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import "testing"

const XGO_VERSION = ""

func TestListStdlib(t *testing.T) {
	if XGO_VERSION == "" {
		t.Fatalf("fail")
	}
}
