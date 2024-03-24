//go:build !windows
// +build !windows

package test

import (
	"strings"
	"testing"
)

// this test is specifically for go1.17,go1.18,go1.19
// when trap not disabled for golang.org/x/sys/unix, the following error happens:
//     golang.org/x/sys/unix.syscall_syscall9·f: relocation target syscall.syscall9 not defined

// go test -run TestGolangXSysCall -v ./test
func TestGolangXSysCall(t *testing.T) {
	t.Parallel()
	output, err := buildAndRunOutputArgs([]string{"./"}, buildAndOutputOptions{
		projectDir: "./testdata/golang_x_sys",
	})
	if err != nil {
		t.Fatal(err)
	}
	expectPrefix := "syscall: 0x"
	if !strings.HasPrefix(output, expectPrefix) {
		t.Fatalf("expect output prefix %q, actual:%q", expectPrefix, output)
	}
}
