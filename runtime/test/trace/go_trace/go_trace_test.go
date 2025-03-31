package go_trace_test

import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/trace/signal"
	"github.com/xhd2015/xgo/runtime/trap/stack_model"
)

func hello(s string) string {
	return "hello " + s
}

func runHello() (interface{}, error) {
	hello("before")
	go hello("world")
	// enough time for the goroutine to finish
	time.Sleep(100 * time.Millisecond)
	hello("after")
	return nil, nil
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

func TestGoTraceSync(t *testing.T) {
	var stack stack_model.IStack
	signal.StartXgoTrace(signal.StartXgoTraceConfig{
		OnFinish: func(s stack_model.IStack) {
			stack = s
		},
	}, nil, runHello)

	json, err := stack.JSON()
	if err != nil {
		t.Fatalf("failed to get stack json: %v", err)
	}
	stackJSON := string(json)
	if !strings.Contains(stackJSON, `"Name":"go"`) {
		t.Fatalf("stack does not contain go")
	}
	t.Logf("stack: %s", string(json))
}

func TestGoTraceAsync(t *testing.T) {
	var stack stack_model.IStack
	signal.StartXgoTrace(signal.StartXgoTraceConfig{
		OnFinish: func(s stack_model.IStack) {
			stack = s
		},
	}, nil, runHelloAsync)

	json, err := stack.JSON()
	if err != nil {
		t.Fatalf("failed to get stack json: %v", err)
	}
	stackJSON := string(json)
	if !strings.Contains(stackJSON, `"Name":"go (running)"`) {
		t.Fatalf("stack does not contain go")
	}
	t.Logf("stack: %s", string(json))
}
