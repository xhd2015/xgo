package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/debug/util"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func TestTracePanicPeek(t *testing.T) {
	var buf bytes.Buffer

	var traceData []byte
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			var err error
			traceData, err = stack.JSON()
			if err != nil {
				t.Fatal(err)
			}
		},
	}, nil, func() (interface{}, error) {
		run(&buf)
		return nil, nil
	})

	output := buf.String()
	// t.Logf("output: %s", s)
	expected := "call: main\ncall: Work\ncall: doWork\nmain panic: Work panic: doWork panic"
	if output != expected {
		t.Fatalf("expect program output: %s, actual: %q", expected, output)
	}

	// t.Logf("traceData: %s", traceData)
	traceFileName := t.Name() + ".json"
	err := os.WriteFile(traceFileName, traceData, 0755)
	if err != nil {
		t.Fatal(err)
	}
	expectTraceSequence := []string{
		"{",
		`"Name":"run",`,
		`"Name":"Work",`,

		// caller error marshalled before callees
		`"Error":"Work panic: doWork panic",`,
		`"Name":"doWork",`,
		`"Error":"doWork panic",`,
		"}",
	}
	err = util.CheckSequence(string(traceData), expectTraceSequence)
	if err != nil {
		t.Logf("traceData: %s", string(traceData))
		stat, statErr := os.Stat(traceFileName)
		if statErr != nil {
			t.Logf("traceFile statErr: %v", statErr)
		} else {
			t.Logf("traceFile size: %d", stat.Size())
		}
		t.Fatalf("%v", err)
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
	doWork(w)
}

func doWork(w io.Writer) {
	fmt.Fprintf(w, "call: doWork\n")
	panic("doWork panic")
}
