// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code
//
// usage:
//  go run -tags dev ./cmd/xgo test --project-dir runtime/test/debug
//  go run -tags dev ./cmd/xgo test --debug-compile --project-dir runtime/test/debug

package debug

import (
	"errors"
	"net"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestDialer(t *testing.T) {
	var mocked bool
	// mock.Patch(http.Get, func(url string) (resp *http.Response, err error) {
	// 	mocked = true
	// 	return nil, errors.New("dial error")
	// })
	mock.Patch((*net.Dialer).Dial, func(_ *net.Dialer, network, address string) (net.Conn, error) {
		mocked = true
		return nil, errors.New("dial error")
	})
	dialer := net.Dialer{}
	dialer.Dial("", "")
	if !mocked {
		t.Fatalf("expected mocked, actually not")
	}
}
