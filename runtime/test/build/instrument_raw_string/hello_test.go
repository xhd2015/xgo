package instrument_raw_string

import "testing"

func TestHello(t *testing.T) {
	if Hello() != "hello" {
		t.Fatal("unexpected result")
	}
}
