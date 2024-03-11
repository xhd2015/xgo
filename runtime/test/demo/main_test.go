package main

import "testing"

func TestTrace(t *testing.T) {
	// ./main.go:7:35: append (built-in) must be called
	// fmt.Printf("what is append: %v", append)
	// a := append
	// _ = a
	// trace("hello", "world")
	trace("hello")
}
