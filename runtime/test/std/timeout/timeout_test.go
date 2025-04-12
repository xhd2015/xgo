package timeout

import (
	"testing"
	"time"
)

// go run ./script/run-test --name timeout --debug -v
func TestTimeout(t *testing.T) {
	time.Sleep(600 * time.Millisecond)

	t.Errorf("this test will error after 600ms, however it should be captured earlier by the test runner and gets ignored")
}
