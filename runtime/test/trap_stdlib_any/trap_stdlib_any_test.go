package trap_stdlib_any

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

// build with --trap-stdlib

func TestTrapStdlibUserHomeDir(t *testing.T) {
	var haveCalledMock bool
	mock.Patch(os.UserHomeDir, func() (string, error) {
		haveCalledMock = true
		return "mock", nil
	})

	homeDir, err := os.UserHomeDir()
	if !haveCalledMock {
		t.Fatalf("mock not called")
	}
	if err != nil {
		t.Fatal(err)
	}
	if homeDir != "mock" {
		t.Fatalf("expect homeDir to be %q, actual: %q", "mock", homeDir)
	}
}

func TestTrapStdlibFuncs(t *testing.T) {
	tests := []struct {
		name string

		fn   interface{}
		call func() error
	}{
		{
			name: "filepath.Join",
			fn:   filepath.Join,
			call: func() error {
				filepath.Join("test")
				return nil
			},
		},
		{
			name: "os.MkdirAll",
			fn:   os.MkdirAll,
			call: func() error {
				return os.MkdirAll("tmp", 0755)
			},
		},
		{
			name: "io.ReadAll",
			fn:   io.ReadAll,
			call: func() error {
				_, err := io.ReadAll(nil)
				return err
			},
		},
		{
			name: "http.Get",
			fn:   http.Get,
			call: func() error {
				_, err := http.Get("test")
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTrapStdlib(t, tt.fn, tt.call)
		})
	}
}

func testTrapStdlib(t *testing.T, fn interface{}, call func() error) {
	var haveCalledMock bool
	mock.Mock(fn, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		haveCalledMock = true
		return nil
	})

	err := call()
	if !haveCalledMock {
		t.Fatalf("mock not called")
	}
	if err != nil {
		t.Fatal(err)
	}
}
