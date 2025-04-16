package runtime

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
)

func MarshalNoError(v interface{}) (result []byte) {
	g := GetG()
	if !g.looseJsonMarshaling {
		g.looseJsonMarshaling = true
		defer func() {
			g.looseJsonMarshaling = false
		}()
	}
	var err error
	defer func() {
		var stackTrace []byte
		if e := recover(); e != nil {
			stackTrace = debug.Stack()
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
		var qstackTrace string
		if len(stackTrace) > 0 {
			qstackTrace = fmt.Sprintf(", %q: %q", "stackTrace", string(stackTrace))
		}
		if err != nil {
			result = []byte(fmt.Sprintf("{%q: %q%s}", "error", err.Error(), qstackTrace))
		}
	}()
	result, err = json.Marshal(v)
	return
}
