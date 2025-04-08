package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/debug/util"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

var a int

func TestTracePanicPeek(t *testing.T) {
	var buf bytes.Buffer

	var traceData []byte
	var traceStack *stack_model.Stack
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			var err error
			traceData, err = stack.JSON()
			if err != nil {
				t.Fatal(err)
			}
			traceStack = stack.Data()
		},
	}, nil, func() (interface{}, error) {
		run(&buf)
		return nil, nil
	})

	output := buf.String()
	// t.Logf("output: %s", s)
	expected := `
call: main
call: Work
call: doWorkBypass
call: doWork a=0
main panic: Work panic: doWork panic
`
	expected = strings.TrimSpace(expected)
	if output != expected {
		t.Fatalf("expect program output: %s, actual: %q", expected, output)
	}

	var hasErr bool
	if bytes.Contains(traceData, []byte("running")) {
		t.Errorf("expect traceData not to contain 'running'")
		hasErr = true
	}

	// t.Logf("traceData: %s", traceData)
	traceFileName := t.Name() + ".json"
	err := os.WriteFile(traceFileName, traceData, 0755)
	if err != nil {
		t.Fatal(err)
	}
	expectStack := `
Trace
 run
  Work
   doWorkBypass
    doWork
     a
`
	expectStack = strings.TrimSpace(expectStack)
	stackBrief := util.BriefStack(traceStack)
	if stackBrief != expectStack {
		hasErr = true
		t.Errorf("expect stack: %s, actual: %s", expectStack, stackBrief)
	}

	if hasErr {
		t.Logf("traceData: %s", string(traceData))
		stat, statErr := os.Stat(traceFileName)
		if statErr != nil {
			t.Logf("traceFile statErr: %v", statErr)
		} else {
			t.Logf("traceFile size: %d", stat.Size())
		}
	}
}

func run(w io.Writer) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(w, "main panic: %v", e)
		}
	}()
	fmt.Fprintf(w, "call: main\n")
	Work(w)
}
func Work(w io.Writer) {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("Work panic: %v", e))
		}
	}()
	fmt.Fprintf(w, "call: Work\n")
	doWorkBypass(w)
}

func doWorkBypass(w io.Writer) {
	fmt.Fprintf(w, "call: doWorkBypass\n")
	doWork(w)
}

func doWork(w io.Writer) {
	fmt.Fprintf(w, "call: doWork a=%d\n", a)
	panic("doWork panic")
}
