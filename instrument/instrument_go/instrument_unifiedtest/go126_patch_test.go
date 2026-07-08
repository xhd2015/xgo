package instrument_unifiedtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

const testGoStubGo126 = `package test

import "context"

func runTest(ctx context.Context,
) {
	pkgs = load.PackagesAndErrors(moduleLoaderState, ctx,
		load.PackageOpts{},
		nil,
	)
}

func (r *runTestActor) Act() {
	cmd.Dir = a.Package.Dir
}
`

const loadTestStubGo126 = `package load

import "context"

func TestPackagesAndErrors(loaderstate *modload.State, ctx context.Context,
	t *testing.T,
) {
	_, err := formatTestmain(t)
}

func loadTestFuncs(ptest *Package) {
}
`

const test2jsonStubGo126 = `package test2json

func (c *Converter) writeEvent(e *event) {
	js, err := json.Marshal(e)
}
`

func writeGo126Goroot(t *testing.T) string {
	t.Helper()
	goroot := t.TempDir()
	dirs := []struct {
		dir  string
		file string
		body string
	}{
		{"src/cmd/go/internal/test", "test.go", testGoStubGo126},
		{"src/cmd/go/internal/load", "test.go", loadTestStubGo126},
		{"src/cmd/internal/test2json", "test2json.go", test2jsonStubGo126},
	}
	for _, d := range dirs {
		p := filepath.Join(goroot, d.dir)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, d.file), []byte(d.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return goroot
}

func TestUnifyGo126PatchesModuleLoaderState(t *testing.T) {
	goroot := writeGo126Goroot(t)
	goVer := &goinfo.GoVersion{Major: 1, Minor: 26}

	if err := Unify(goroot, goVer); err != nil {
		t.Fatalf("Unify go1.26: %v", err)
	}

	testPath := filepath.Join(goroot, "src", "cmd", "go", "internal", "test", "test.go")
	testContent, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(testContent), "xgoModuleLoaderState = moduleLoaderState") {
		t.Fatalf("test.go missing moduleLoaderState assignment:\n%s", testContent)
	}
	if !strings.Contains(string(testContent), "pkgs = xgoUnifyTestPackages(ctx, pkgs)") {
		t.Fatalf("test.go missing unify call:\n%s", testContent)
	}

	unifiedPath := filepath.Join(goroot, "src", "cmd", "go", "internal", "test", "xgo_testunified.go")
	unifiedContent, err := os.ReadFile(unifiedPath)
	if err != nil {
		t.Fatal(err)
	}
	u := string(unifiedContent)
	if !strings.Contains(u, "cmd/go/internal/modload") {
		t.Fatalf("xgo_testunified.go missing modload import:\n%s", u)
	}
	if !strings.Contains(u, "xgoModuleLoaderState *modload.State") {
		t.Fatalf("xgo_testunified.go missing loader state var:\n%s", u)
	}
	if !strings.Contains(u, "load.PackagesAndErrors(xgoModuleLoaderState, ctx,") {
		t.Fatalf("xgo_testunified.go missing loader state arg:\n%s", u)
	}

	loadPath := filepath.Join(goroot, "src", "cmd", "go", "internal", "load", "test.go")
	loadContent, err := os.ReadFile(loadPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(loadContent), "func TestPackagesAndErrors(loaderstate *modload.State, ctx context.Context,") {
		t.Fatalf("load/test.go anchor not preserved:\n%s", loadContent)
	}
	if !strings.Contains(string(loadContent), "XgoAfterGenerateTestMain") {
		t.Fatalf("load/test.go missing testmain hook:\n%s", loadContent)
	}
}