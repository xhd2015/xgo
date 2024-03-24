package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

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
		traceData, err = json.Marshal(root.Export())
		if err != nil {
			t.Fatal(err)
		}
	}).Collect(func() {
		run(&buf)
	})

	s := buf.String()
	t.Logf("buf: %s", s)

	t.Logf("traceData: %s", traceData)
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
