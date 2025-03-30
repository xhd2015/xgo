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

	recorders := stack.getVarRecordHandlers(ptr)
	var postRecorders []func()
	for _, recorder := range recorders {
		var data interface{}
		if recorder.pre != nil {
			data, _ = recorder.pre(name, res)
		}
		if recorder.post != nil {
			postRecorders = append(postRecorders, func() {
				recorder.post(name, res, data)
			})
		}
	}

	mock := stack.getLastVarMock(ptr)
	if mock != nil {
		mock(name, res)
	}
	for _, recorder := range postRecorders {
		recorder()
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

	recorders := stack.getVarPtrRecordHandlers(ptr)
	var postRecorders []func()
	for _, recorder := range recorders {
		var data interface{}
		if recorder.pre != nil {
			data, _ = recorder.pre(name, res)
		}
		if recorder.post != nil {
			postRecorders = append(postRecorders, func() {
				recorder.post(name, res, data)
			})
		}
	}

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

	for _, recorder := range postRecorders {
		recorder()
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
