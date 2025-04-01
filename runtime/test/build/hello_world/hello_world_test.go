package main

import "testing"

// go install github.com/xhd2015/kool@latest
// kool with-go1.19.13 go run -tags dev ./cmd/xgo test --log-debug --project-dir runtime/test ./hello_world
func TestHello(t *testing.T) {
	t.Logf("hello world!")
}
