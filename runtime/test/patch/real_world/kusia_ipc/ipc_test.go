package kusia_ipc

import (
	"testing"

	"github.com/apache/arrow/go/v13/arrow/flight"
	"github.com/xhd2015/xgo/runtime/mock"
)

// source: https://github.com/secretflow/kuscia
func TestFlightStreamToDataProxyContentBinary_ErrorFormat(t *testing.T) {

	reader2222 := &flight.Reader{}
	times := 0
	mock.Patch(reader2222.Reader.Next, func() bool {
		times++
		// first return true, second return false
		return times == 1
	})

}
