package trace_sleep

import (
	"runtime"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/test/debug/util"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func TestTraceSleep(t *testing.T) {
	var record1 stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			record1 = stack
		},
	}, nil, func() (interface{}, error) {
		doSleep(10 * time.Millisecond)
		return nil, nil
	})

	var record2 stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			record2 = stack
		},
	}, nil, func() (interface{}, error) {
		doSleep(200 * time.Millisecond)
		return nil, nil
	})
	cost1 := util.GetCostNs(record1.Data(), "Trace", "doSleep")

	cost2 := util.GetCostNs(record2.Data(), "Trace", "doSleep")

	cost1Dur := time.Duration(cost1) * time.Nanosecond
	cost2Dur := time.Duration(cost2) * time.Nanosecond

	const BUF = 10 * time.Millisecond
	const BUF_WINDOWS = 20 * time.Millisecond
	if cost1Dur < 10*time.Millisecond {
		t.Errorf("expect cost1 to be greater than 10ms: %v", cost1Dur)
	}
	// leave a 10ms buffer
	if cost1Dur > 10*time.Millisecond+BUF {
		// windows is the slow one
		// trace_sleep_test.go:45: expect cost1 to be less than 10ms+BUF: 20.9081ms > 20ms
		if runtime.GOOS == "windows" {
			t.Logf("expect cost1 to be less than 10ms+BUF: %v > %v", cost1Dur, 10*time.Millisecond+BUF)
			if cost1Dur > 10*time.Millisecond+BUF_WINDOWS {
				t.Errorf("expect cost1 to be less than 10ms+BUF_WINDOWS: %v > %v", cost1Dur, 10*time.Millisecond+BUF_WINDOWS)
			}
		} else {
			t.Errorf("expect cost1 to be less than 10ms+BUF: %v > %v", cost1Dur, 10*time.Millisecond+BUF)
		}
	}
	if cost2Dur < 200*time.Millisecond {
		t.Errorf("expect cost2 to be greater than 200ms: %v", cost2Dur)
	}
	if cost2Dur > 200*time.Millisecond+BUF {
		t.Errorf("expect cost2 to be less than 200ms+BUF: %v > %v", cost2Dur, 200*time.Millisecond+BUF)
	}

}

func doSleep(n time.Duration) {
	time.Sleep(n)
}
