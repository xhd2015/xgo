package prog_arg

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// run example:
//
//	go run -tags dev ./cmd/xgo e --project-dir cmd/xgo/test-explorer/testdata/prog_arg
func TestProgArg(t *testing.T) {
	// NOTE: args example:
	//   [-test.paniconexit0 -test.timeout=10m0s -test.v=true challenge]
	args := os.Args[1:]

	var progArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-test.") {
			continue
		}
		progArgs = append(progArgs, arg)
	}

	if !reflect.DeepEqual(progArgs, []string{"challenge"}) {
		t.Logf("expect args contain challenge, actual: %v", progArgs)
	}
}
