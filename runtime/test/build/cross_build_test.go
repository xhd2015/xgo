package xgo_integration

import (
	"testing"
)

// go test ./test/xgo_integration/ -v -run TestCrossBuild
func TestCrossBuild(t *testing.T) {
	// xgo test -c ./simple

	// _, err := exec.LookPath("xgo")
	// if err != nil {
	// 	t.Skipf("missing: %v", err)
	// }
	// err = cmd.Env([]string{"GOOS=windows", "GOARCH=amd64"}).Debug().Dir("./testdata/build_simple").Run("xgo", "test", "--log-debug", "-c", "-o", "/dev/null")
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
