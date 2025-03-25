package trace_runtime

import (
	"fmt"
	"reflect"
)

type mockHolder struct {
	wantRecvPtr interface{}
	mock        func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{})
}

// PushMockHandler pushes a mock handler to the stack.
// The returned function can be used to pop the mock.
// If the mock is not popped, it will affect even after
// the caller returned.
func PushMockHandler(pc uintptr, recvPtr interface{}, mock func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{})) func() {
	stack := GetOrAttachStack()
	if stack.mock == nil {
		stack.mock = map[uintptr][]*mockHolder{}
	}
	h := &mockHolder{wantRecvPtr: recvPtr, mock: mock}
	stack.mock[pc] = append(stack.mock[pc], h)
	return func() {
		list := stack.mock[pc]
		n := len(list)
		if list[n-1] == h {
			stack.mock[pc] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				stack.mock[pc] = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop mock not found, check if the mock is already popped earlier"))
	}
}

func (c *Stack) getLastMock(pc uintptr) (recvPtr interface{}, mock func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{})) {
	mockList := c.mock[pc]
	if len(mockList) == 0 {
		return nil, nil
	}
	m := mockList[len(mockList)-1]
	return m.wantRecvPtr, m.mock
}

func sameReceiver(recvPtr interface{}, actRecvPtr interface{}) bool {
	// assume both are non-nil
	recvPtrVal := reflect.ValueOf(recvPtr)
	actRecvPtrVal := reflect.ValueOf(actRecvPtr)
	return recvPtrVal.Elem().Interface() == actRecvPtrVal.Elem().Interface()
}
