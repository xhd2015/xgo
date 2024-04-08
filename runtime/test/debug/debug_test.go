// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"
)

const good = 2
const reason = "test"

func TestPatchConstOperationShouldCompileAndSkipMock(t *testing.T) {
	reasons := getReasons("good")
	if len(reasons) != 2 || reasons[0] != "ok" || reasons[1] != "good" {
		t.Fatalf("bad reason: %v", reasons)
	}

	getReasons2 := func(good string) (reason []string) {
		reason = append(reason, "ok")
		reason = append(reason, good)
		return
	}
	reasons2 := getReasons2("good")
	if len(reasons2) != 2 || reasons2[0] != "ok" || reasons2[1] != "good" {
		t.Fatalf("bad reason2: %v", reasons2)
	}
}

func getReasons(good string) (reason []string) {
	reason = append(reason, "ok")
	reason = append(reason, good)
	return
}
