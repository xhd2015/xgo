package prog_arg

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// run example:
//
//	go run -tags dev ./cmd/xgo e --project-dir cmd/xgo/test-explorer/testdata/mock_rules
func TestProgArg(t *testing.T) {
	var panicErr interface{}
	func() {
		defer func() {
			panicErr = recover()
		}()
		mock.Patch(http.Get, func(url string) (*http.Response, error) {
			return nil, nil
		})
	}()

	if panicErr == nil {
		t.Fatalf("expect mock http.Get panic, actually not")
	}

	msg := fmt.Sprint(panicErr)
	expect := "failed to setup mock for: net/http.Get"
	if msg != expect {
		t.Fatalf("expect panic %q, actual: %q", expect, msg)
	}
}
