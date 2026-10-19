package targeted_packages

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// xgo instruments only the main module packages that the command builds.
// Testing one package must still instrument everything its tests can mock:
// its dependencies, packages imported only by its tests, and their
// dependencies, and packages that its external tests import which import
// it in turn (helperdep and consumer are mocked by name, which only works
// if the package itself was instrumented).
func TestSinglePackageMocksDepsAndTestOnlyImports(t *testing.T) {
	root := findXgoRepoRoot(t)
	modDir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(modDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/targeted\n\ngo 1.18\n\n"+
		"require github.com/xhd2015/xgo/runtime v0.0.0\n\n"+
		"replace github.com/xhd2015/xgo/runtime => "+filepath.ToSlash(filepath.Join(root, "runtime"))+"\n")
	write("dep/dep.go", "package dep\n\nfunc Compute() int { return 1 }\n")
	write("helperdep/helperdep.go", "package helperdep\n\nfunc Seed() string { return \"seed\" }\n")
	write("testutil/testutil.go", "package testutil\n\nimport \"example.com/targeted/helperdep\"\n\n"+
		"func Fixture() string { return \"fixture:\" + helperdep.Seed() }\n")
	write("unrelated/unrelated.go", "package unrelated\n\nfunc Unused() int { return 0 }\n")
	// consumer imports svc and only svc's external test imports consumer:
	// go list -test reports it only as a variant built for svc's test
	write("consumer/consumer.go", "package consumer\n\nimport \"example.com/targeted/svc\"\n\n"+
		"func Use() int { return svc.Total() * 2 }\n")
	write("svc/svc_ext_test.go", `package svc_test

import (
	"testing"

	"example.com/targeted/consumer"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestConsumer(t *testing.T) {
	mock.PatchByName("example.com/targeted/consumer", "Use", func() int { return 42 })
	if got := consumer.Use(); got != 42 {
		t.Errorf("Use mocked by name: %d", got)
	}
}
`)
	write("svc/svc.go", "package svc\n\nimport \"example.com/targeted/dep\"\n\n"+
		"func Total() int { return dep.Compute() + 1 }\n")
	write("svc/svc_test.go", `package svc

import (
	"testing"

	"example.com/targeted/dep"
	"example.com/targeted/testutil"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMocks(t *testing.T) {
	mock.Patch(dep.Compute, func() int { return 10 })
	if got := Total(); got != 11 {
		t.Errorf("Total with dep.Compute mocked: %d", got)
	}
	// by name: only works if the package itself was instrumented, as
	// nothing in the test refers to the function
	mock.PatchByName("example.com/targeted/helperdep", "Seed", func() string { return "patched" })
	if got := testutil.Fixture(); got != "fixture:patched" {
		t.Errorf("Fixture with helperdep.Seed mocked: %q", got)
	}
	mock.Patch(testutil.Fixture, func() string { return "mocked" })
	if got := testutil.Fixture(); got != "mocked" {
		t.Errorf("Fixture mocked: %q", got)
	}
}

// only the test files of the package rewrite testLimit
var testLimit = 3

func getTestLimit() int { return testLimit }

func TestMockVarOfTestFile(t *testing.T) {
	mock.Patch(&testLimit, func() int { return 9 })
	if got := getTestLimit(); got != 9 {
		t.Errorf("testLimit mocked: %d", got)
	}
}
`)
	// go list reports "./svc/" as "./svc": it must be targeted all the same
	for _, pattern := range []string{"./svc", "./svc/"} {
		cmd := exec.Command("go", "run", "./cmd/xgo", "test", "--project-dir", modDir, "-count=1", "-timeout=60s", pattern)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GOWORK=off")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			t.Fatalf("nested xgo test %s failed: %v\n%s", pattern, err, buf.String())
		}
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
