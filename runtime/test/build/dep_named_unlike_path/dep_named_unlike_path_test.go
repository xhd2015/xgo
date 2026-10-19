package dep_named_unlike_path

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A variable whose type comes from a dependency outside the main module is
// trapped only if its import resolves to the package name the dependency
// declares (lib), not to the "lib-go" a guess from the import path gives,
// like github.com/openai/openai-go is package openai. Dependencies outside
// the main module are listed, not parsed, before the instrumentation
// resolves references, so their names must come from the listing.
func TestMockVarOfDependencyTypeNamedUnlikePath(t *testing.T) {
	root := findXgoRepoRoot(t)
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("lib-go/go.mod", "module example.com/lib-go\n\ngo 1.18\n")
	write("lib-go/lib.go", "package lib\n\ntype Config struct{ Name string }\n")
	write("app/go.mod", "module example.com/app\n\ngo 1.18\n\n"+
		"require (\n\texample.com/lib-go v0.0.0\n\tgithub.com/xhd2015/xgo/runtime v0.0.0\n)\n\n"+
		"replace example.com/lib-go => ../lib-go\n\n"+
		"replace github.com/xhd2015/xgo/runtime => "+filepath.ToSlash(filepath.Join(root, "runtime"))+"\n")
	write("app/app.go", `package app

import "example.com/lib-go"

var config = lib.Config{Name: "real"}

func Name() string { return config.Name }
`)
	write("app/app_test.go", `package app

import (
	"testing"

	"example.com/lib-go"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMockVar(t *testing.T) {
	mock.Patch(&config, func() lib.Config { return lib.Config{Name: "mocked"} })
	if got := Name(); got != "mocked" {
		t.Fatalf("config not mocked: %q", got)
	}
}
`)
	cmd := exec.Command("go", "run", "./cmd/xgo", "test", "--project-dir", filepath.Join(dir, "app"), "-count=1", "-timeout=60s", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		t.Fatalf("nested xgo test failed: %v\n%s", err, buf.String())
	}
}

func findXgoRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.HasPrefix(bytes.TrimSpace(data), []byte("module github.com/xhd2015/xgo\n")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("xgo repo root not found")
		}
		dir = parent
	}
}
