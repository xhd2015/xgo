package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/util"
	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Enable()
}

func TestTracePanicPeek(t *testing.T) {
	var buf bytes.Buffer

	var traceData []byte
	trace.Options().OnComplete(func(root *trace.Root) {
		var err error
		traceData, err = trace.MarshalAnyJSON(root.Export(nil))
		if err != nil {
			t.Fatal(err)
		}
	}).Collect(func() {
		run(&buf)
	})

	output := buf.String()
	// t.Logf("output: %s", s)
	expected := "call: main\ncall: Work\ncall: doWork\nmain panic: Work panic: doWork panic"
	if output != expected {
		t.Fatalf("expect program output: %s, actual: %q", expected, output)
	}

	// t.Logf("traceData: %s", traceData)
	expectTraceSequence := []string{
		"{",
		`"Name":"run",`,
		`"Name":"Work",`,
		`"Name":"doWork",`,
		`"Error":"panic: doWork panic",`,
		`"Error":"Work panic: doWork panic",`,
		"}",
	}
	err := util.CheckSequence(string(traceData), expectTraceSequence)
	if err != nil {
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
