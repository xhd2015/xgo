package mock_closuer

import "testing"

// see https://github.com/xhd2015/xgo/issues/13
func TestClosuerWithIfFalse(t *testing.T) {
	// assert this function
	var a int
	if false {
		func() {
			a = 10
		}()
	}
	if a != 0 {
		t.Fatalf("expect a to be 0, actual: %d", a)
	}
}
func TestClosuerWithIfTrue(t *testing.T) {
	var a int
	if true {
		func() {
			a = 10
		}()
	}
	if a != 10 {
		t.Fatalf("expect a to be 10, actual: %d", a)
	}
}
