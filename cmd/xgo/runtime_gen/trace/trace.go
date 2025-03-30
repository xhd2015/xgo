package trace

// trace v2's core functionality are all implemented in trap package
import (
	"github.com/xhd2015/xgo/runtime/trap"
)

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

func RecordResult(v interface{}, post interface{}) func() {
	return trap.PushRecorder(v, nil, post)
}

func RecordInterceptor(v interface{}, pre trap.PreRecordInterceptor, post trap.PostRecordInterceptor) func() {
	return trap.PushRecorderInterceptor(v, pre, post)
}
