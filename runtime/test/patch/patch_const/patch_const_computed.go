//go:build go1.20
// +build go1.20

package patch_const

import "testing"

const a = iota

const (
	b0 = iota
	b1
)

const c = 1<<2 + 3

func TestConstComputedShouldCompile(t *testing.T) {
	var testA int64 = a
	if testA != 0 {
		t.Fatalf("expect testA to be %d, actual: %d", 0, testA)
	}

	var testB0 int64 = b0
	if testB0 != 0 {
		t.Fatalf("expect testB0 to be %d, actual: %d", 0, testB0)
	}

	var testB1 int64 = b1
	if testB1 != 0 {
		t.Fatalf("expect testB1 to be %d, actual: %d", 1, testB1)
	}

	var testC int64 = c
	if testC != 7 {
		t.Fatalf("expect testC to be %d, actual: %d", 7, testC)
	}
}
