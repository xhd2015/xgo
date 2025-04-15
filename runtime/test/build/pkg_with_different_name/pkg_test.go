package pkg_with_different_name

import (
	"testing"

	"github.com/rwynn/gtm/v2"
	"github.com/xhd2015/xgo/runtime/trace"
	"go.mongodb.org/mongo-driver/mongo"
)

// see https://github.com/xhd2015/xgo/issues/317
func TestPkg(t *testing.T) {
	var called bool
	trace.RecordCall(gtm.Start, func(client *mongo.Client, o *gtm.Options, res *gtm.OpCtx) {
		called = true
		panic("test")
	})
	func() {
		defer func() {
			if e := recover(); e != nil {
				t.Logf("gtm.Start panicked: %v", e)
			}
		}()
		gtm.Start(nil, nil)
	}()
	if !called {
		t.Fatal("gtm.Start not called")
	}
}
