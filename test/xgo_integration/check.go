package xgo_integration

import (
	"os"
	"os/exec"
	"sync"
	"testing"
)

// must have xgo installed first
// otherwise skipped

var checkXgoOnce sync.Once
var checkXgoErr error

func checkXgo(t *testing.T) {
	checkXgoOnce.Do(func() {
		_, checkXgoErr = exec.LookPath("xgo")
	})
	if checkXgoErr == nil {
		return
	}
	t.Skipf("skipped because xgo not installed: %v", checkXgoErr)
}

var envLock sync.Mutex

func withPathEnv(val string, f func()) {
	envLock.Lock()
	defer envLock.Unlock()
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", val)
	defer os.Setenv("PATH", oldPath)
	f()
}
