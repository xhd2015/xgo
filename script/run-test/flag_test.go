package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestAppendShortIfFast(t *testing.T) {
	tests := []struct {
		name       string
		remainArgs []string
		want       []string
	}{
		{name: "no existing -short", remainArgs: []string{"-v"}, want: []string{"-v", "-short"}},
		{name: "with existing -short", remainArgs: []string{"-v", "-short"}, want: []string{"-v", "-short"}},
		{name: "empty remain args", remainArgs: nil, want: []string{"-short"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendShortIfFast(tt.remainArgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParseUseFilePatchesFlag(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		want    *bool
		wantErr bool
	}{
		{name: "true", val: "true", want: ptrBool(true)},
		{name: "false", val: "false", want: ptrBool(false)},
		{name: "empty (bare flag)", val: "", want: ptrBool(true)},
		{name: "invalid", val: "bad", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUseFilePatchesFlag(tt.val)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if *got != *tt.want {
				t.Fatalf("expected %v, got %v", *tt.want, *got)
			}
		})
	}
}

func TestBuildReproduceCmd(t *testing.T) {
	goVersion := &goinfo.GoVersion{Major: 1, Minor: 25, Patch: 10}
	testPackages := []string{"./cmd/exec_tool"}

	rf := reproduceFlags{
		installXgo:   true,
		withSetup:    true,
		resetInstr:   true,
		logDebug:     true,
		remainArgs:   []string{"-v"},
	}

	cmd := buildReproduceCmd(goVersion, testPackages, rf)

	if !strings.Contains(cmd, "go run ./script/run-test") {
		t.Fatalf("expected 'go run ./script/run-test' prefix, got %q", cmd)
	}
	if !strings.Contains(cmd, "--include go1.25.10") {
		t.Fatalf("expected '--include go1.25.10', got %q", cmd)
	}
	if !strings.Contains(cmd, "--install-xgo") {
		t.Fatalf("expected '--install-xgo', got %q", cmd)
	}
	if !strings.Contains(cmd, "--with-setup") {
		t.Fatalf("expected '--with-setup', got %q", cmd)
	}
	if !strings.Contains(cmd, "--reset-instrument") {
		t.Fatalf("expected '--reset-instrument', got %q", cmd)
	}
	if !strings.Contains(cmd, "--log-debug") {
		t.Fatalf("expected '--log-debug', got %q", cmd)
	}
	if !strings.Contains(cmd, "-v") {
		t.Fatalf("expected '-v', got %q", cmd)
	}
	if !strings.Contains(cmd, "./cmd/exec_tool") {
		t.Fatalf("expected './cmd/exec_tool', got %q", cmd)
	}
	// verify flags that are NOT set don't appear
	if strings.Contains(cmd, "--debug") && !strings.Contains(cmd, "--debug-xgo") {
		t.Fatalf("unexpected '--debug' in %q", cmd)
	}
	if strings.Contains(cmd, "-race") {
		t.Fatalf("unexpected '-race' in %q", cmd)
	}
}

func TestBuildReproduceCmdMinimal(t *testing.T) {
	goVersion := &goinfo.GoVersion{Major: 1, Minor: 21, Patch: 8}
	testPackages := []string{"./..."}

	rf := reproduceFlags{}

	cmd := buildReproduceCmd(goVersion, testPackages, rf)

	if !strings.Contains(cmd, "go run ./script/run-test") {
		t.Fatalf("expected 'go run ./script/run-test' prefix, got %q", cmd)
	}
	if !strings.Contains(cmd, "--include go1.21.8") {
		t.Fatalf("expected '--include go1.21.8', got %q", cmd)
	}
	if strings.Contains(cmd, "--install-xgo") {
		t.Fatalf("unexpected '--install-xgo' in minimal cmd: %q", cmd)
	}
	if !strings.Contains(cmd, "./...") {
		t.Fatalf("expected './...' package, got %q", cmd)
	}
}

