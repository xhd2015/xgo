package race_export_trap_all_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func hello(s string) string {
	return "hello " + s
}

func helloSlow(s string) string {
	// Keep the child in trap/instrumented work while parent may export.
	time.Sleep(30 * time.Millisecond)
	return "hello " + s
}

// TestGoTraceExportAcrossGoNoRace_trap_all is a green canary under
// -race --trap-all (no --xgo-race-safe).
//
// What it guards beyond the plain -race race_export canary:
//   - Concurrent export vs child Push while trap can re-enter (trap-all).
//   - Export paths that still call instrumentable time methods
//     (Format / UnixNano) must not hang or DATA RACE under full trap.
//
// Expected PASS with go child present and zero DATA RACE.
func TestGoTraceExportAcrossGoNoRace_trap_all(t *testing.T) {
	var stack stack_model.IStack
	var wg sync.WaitGroup
	const n = 16

	_, _ = trace.Trace(trace.Config{
		OnFinish: func(s stack_model.IStack) {
			stack = s
		},
	}, nil, func() (interface{}, error) {
		hello("before")
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				helloSlow("world")
			}()
		}
		// Do not wg.Wait() before Trace returns: OnFinish export must overlap
		// with still-running children (R1 concurrency).
		hello("after")
		time.Sleep(5 * time.Millisecond)
		return nil, nil
	})

	wg.Wait()

	if stack == nil {
		t.Fatal("stack is nil")
	}
	json, err := stack.JSON()
	if err != nil {
		t.Fatalf("stack JSON: %v", err)
	}
	// Name may be "go" or "go (running)".
	if !strings.Contains(string(json), `"Name":"go`) {
		t.Fatalf("expected go child in stack export, got:\n%s", json)
	}
}
