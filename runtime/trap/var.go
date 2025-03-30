package trap

import (
	"encoding/json"
	"reflect"
	"runtime"
	"time"
)

const SKIP = 2

func trapVar(name string, varAddr interface{}, res interface{}) {
	begin := time.Now()
	stack := GetStack()
	if stack == nil {
		return
	}

	ptr := reflect.ValueOf(varAddr).Pointer()
	mock := stack.getLastVarMock(ptr)
	if mock != nil {
		mock(name, res)
	}
	if !stack.hasStartedTracing {
		return
	}
	_, file, line, _ := runtime.Caller(SKIP + 1)
	cur := stack.newStackEntry(begin, name)
	cur.File = file
	cur.Line = line
	cur.Results = json.RawMessage(marshalNoError(res))
	stack.Top = stack.push(cur)
	cur.EndNs = time.Now().UnixNano() - stack.Begin.UnixNano()
}

func trapVarPtr(name string, varAddr interface{}, res interface{}) {
	begin := time.Now()
	stack := GetStack()
	if stack == nil {
		return
	}

	mockRes := res
	ptr := reflect.ValueOf(varAddr).Pointer()
	mock := stack.getLastVarPtrMock(ptr)
	if mock == nil {
		mock = stack.getLastVarMock(ptr)
		if mock != nil {
			// res: **T
			// mockRes: *T
			mockRes = reflect.ValueOf(res).Elem().Interface()
		}
	}

	if mock != nil {
		mock(name, mockRes)
	}
	if !stack.hasStartedTracing {
		return
	}

	_, file, line, _ := runtime.Caller(SKIP + 1)
	cur := stack.newStackEntry(begin, name)
	cur.File = file
	cur.Line = line
	cur.Results = json.RawMessage(marshalNoError(res))
	stack.Top = stack.push(cur)
	cur.EndNs = time.Now().UnixNano() - stack.Begin.UnixNano()
}
