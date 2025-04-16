package trace

// trace v2's core functionality are all implemented in trap package
import (
	"github.com/xhd2015/xgo/runtime/internal/trap"
)

// Record setup a recorder for v, with pre-hook and post-hook
func Record(v interface{}, pre interface{}, post interface{}) func() {
	return trap.PushRecorder(v, pre, post)
}

// RecordCall is a convenience function for recording a call to a function
// with a pre-hook.
// a practical example would be:
//
//	func SetupLog(t *testing.T) {
//		trace.RecordCall(log.Errorf, func(ctx context.Context, format string, params []interface{}) {
//			t.Logf("ERROR "+format, params...)
//		})
//	}
func RecordCall(v interface{}, pre interface{}) func() {
	return trap.PushRecorder(v, pre, nil)
}

// RecordResult is a convenience function for recording a result from a function
func RecordResult(v interface{}, post interface{}) func() {
	return trap.PushRecorder(v, nil, post)
}
