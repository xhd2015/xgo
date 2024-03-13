package test

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// go test -run TestLongFuncNoSplitShouldNotCompileWithDebugFlags -v ./test
func TestLongFuncNoSplitShouldNotCompileWithDebugFlags(t *testing.T) {
	var errBuf bytes.Buffer
	_, buildErr := buildAndRunOutputArgs([]string{"-gcflags=all=-N -l", "./testdata/nosplit/long_func_overflow.go"}, buildAndOutputOptions{
		build: func(args []string) error {
			// use go build
			buildCmd := exec.Command("go", append([]string{"build"}, args...)...)
			buildCmd.Stderr = &errBuf
			buildCmd.Stdout = os.Stdout
			return buildCmd.Run()
		},
	})
	if buildErr == nil {
		t.Fatalf("expect build fail")
	}
	errOutput := errBuf.String()

	// t.Logf("output: %s", errOutput)

	expect := "main.longFunc: nosplit stack over 792 byte limit"
	if !strings.Contains(errOutput, expect) {
		t.Fatalf("expect err output contains: %q, actual not found", expect)
	}
}

// go test -run TestSmallFuncNoSplitShouldCompileWithDebugFlagsWithGo -v ./test
func TestSmallFuncNoSplitShouldCompileWithDebugFlagsWithGo(t *testing.T) {
	output, err := buildAndRunOutputArgs([]string{"-gcflags=all=-N -l", "./testdata/nosplit/small_func_should_compile.go"}, buildAndOutputOptions{
		build: func(args []string) error {
			// use go build
			buildCmd := exec.Command("go", append([]string{"build"}, args...)...)
			buildCmd.Stderr = os.Stderr
			buildCmd.Stdout = os.Stdout
			return buildCmd.Run()
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	expect := "hello nosplit:<nil>\n"
	if output != expect {
		t.Fatalf("expect output: %q, actual: %q", expect, output)
	}
}

// go test -run TestSmallFuncNoSplitShouldCompileWithoutDebugWithXgo -v ./test
func TestSmallFuncNoSplitShouldCompileWithoutDebugWithXgo(t *testing.T) {
	output, err := buildAndRunOutputArgs([]string{"./testdata/nosplit/small_func_should_compile.go"}, buildAndOutputOptions{})
	if err != nil {
		t.Fatal(err)
	}

	expect := "hello nosplit:<nil>\n"
	if output != expect {
		t.Fatalf("expect output: %q, actual: %q", expect, output)
	}
}

// go test -run TestSmallFuncNoSplitShouldCompileWithDebugFlagsWithXgo -v ./test
func TestSmallFuncNoSplitShouldCompileWithDebugFlagsWithXgo(t *testing.T) {
	output, err := buildAndRunOutputArgs([]string{"--", "-gcflags=all=-N -l", "./testdata/nosplit/small_func_should_compile.go"}, buildAndOutputOptions{})
	if err != nil {
		t.Fatal(err)
	}

	expect := "hello nosplit:<nil>\n"
	if output != expect {
		t.Fatalf("expect output: %q, actual: %q", expect, output)
	}
}

// TODO: add test that nosplit are not trapped
