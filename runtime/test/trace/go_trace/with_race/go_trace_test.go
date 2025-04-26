package with_race_test

import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func hello(s string) string {
	return "hello " + s
}

func helloAsync(s string) string {
	time.Sleep(100 * time.Millisecond)
	return "hello " + s
}

func runHelloAsync() (interface{}, error) {
	hello("before")
	go helloAsync("world")
	// enough time for the goroutine to finish
	hello("after")
	return nil, nil
}

func TestGoTraceAsync(t *testing.T) {
	var stack stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(s stack_model.IStack) {
			stack = s
		},
	}, nil, runHelloAsync)

	json, err := stack.JSON()
	if err != nil {
		t.Fatalf("failed to get stack json: %v", err)
	}
	stackJSON := string(json)
	const expectNotContainName = `"Name":"go`
	if strings.Contains(stackJSON, expectNotContainName) {
		t.Errorf("expect not contain, actually stack contains %q", expectNotContainName)
	}
	// t.Logf("stack: %s", string(json))
}
