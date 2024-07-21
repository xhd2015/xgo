package main

import (
	"testing"
)

func TestUnmarshal(t *testing.T) {
	if goMajor > 1 || goMinor > 22 {
		t.Skip("this sonic thing only supports go1.22 as of 2024-07-22")
	}
	unmarshal(nil, nil)
}
