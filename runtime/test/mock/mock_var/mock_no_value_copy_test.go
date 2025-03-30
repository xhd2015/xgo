package mock_var

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var lock sync.Mutex
var count int

func TestMockLockShouldNotCopy(t *testing.T) {
	var buf bytes.Buffer
	cancel := trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			buf.WriteString(fmt.Sprintf("%s\n", f.IdentityName))
			return
		},
	})
	incrLocked()
	cancel()
	if count != 1 {
		t.Fatalf("expect count to be 1,actual: %d", count)
	}
	trapStr := buf.String()

	// notice there is no lock
	expectTrapStr := "incrLocked\ncount\n"
	if trapStr != expectTrapStr {
		t.Fatalf("expect tarp buf: %q, actual: %q", expectTrapStr, trapStr)
	}
}

func incrLocked() {
	lock.Lock()
	count = count + 1
	lock.Unlock()
}

// NOTE: this test case demonstrates a buggy case:
//
//	f := lock.Lock, so lock should not be intercepted
func TestMockLockFuncShouldNotBeMistakenlyCopied(t *testing.T) {
	var buf bytes.Buffer
	cancel := trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			buf.WriteString(fmt.Sprintf("%s\n", f.IdentityName))
			return
		},
	})
	incrLocked_Fn()
	cancel()
	if countFn != 1 {
		t.Fatalf("expect countFn to be 1,actual: %d", countFn)
	}
	trapStr := buf.String()

	// notice there is no lock
	expectTrapStr := "incrLocked_Fn\ncountFn\n"
	if trapStr != expectTrapStr {
		t.Fatalf("expect tarp buf: %q, actual: %q", expectTrapStr, trapStr)
	}
}

var countFn int

// NOTE: if lock is captured, this function will
// panic: fatal error: sync: unlock of unlocked mutex
func incrLocked_Fn() {
	f := lock.Lock
	f()
	countFn = countFn + 1
	lock.Unlock()
}

func TestMockLockFuncParenShouldNotBeMistakenlyCopied(t *testing.T) {
	var buf bytes.Buffer
	cancel := trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			buf.WriteString(fmt.Sprintf("%s\n", f.IdentityName))
			return
		},
	})
	incrLocked_Paren()
	cancel()
	if countParen != 1 {
		t.Fatalf("expect countParen to be 1,actual: %d", countParen)
	}
	trapStr := buf.String()

	// notice there is no lock
	expectTrapStr := "incrLocked_Paren\ncountParen\n"
	if trapStr != expectTrapStr {
		t.Fatalf("expect tarp buf: %q, actual: %q", expectTrapStr, trapStr)
	}
}

var countParen int

// NOTE: if lock is captured, this function will
// panic: fatal error: sync: unlock of unlocked mutex
func incrLocked_Paren() {
	f := (lock).Lock
	f()
	countParen = countParen + 1
	lock.Unlock()
}
