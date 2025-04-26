package debug

import (
	"runtime"
	"testing"
)

func TestDebug(t *testing.T) {
	t.Log("hello")
	cpuLimit := runtime.GOMAXPROCS(0)
	t.Logf("cpuLimit: %d", cpuLimit) // 10
}
