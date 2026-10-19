package concurrent_runs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// xgo runs in the same project, e.g. for different packages, generate
// different overlays, compiler options and an instrumented copy of the xgo
// runtime: one run must not overwrite what the go command of another one
// is reading.
func TestConcurrentRunsInOneProject(t *testing.T) {
	root := findXgoRepoRoot(t)
	tmp := t.TempDir()
	xgo := filepath.Join(tmp, "xgo")
	if runtime.GOOS == "windows" {
		xgo += ".exe"
	}
	build := exec.Command("go", "build", "-o", xgo, "./cmd/xgo")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build xgo: %v\n%s", err, out)
	}

	modDir := filepath.Join(tmp, "mod")
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
	write("go.mod", "module example.com/conc\n\ngo 1.18\n\n"+
		"require github.com/xhd2015/xgo/runtime v0.0.0\n\n"+
		"replace github.com/xhd2015/xgo/runtime => "+filepath.ToSlash(filepath.Join(root, "runtime"))+"\n")
	pkgs := []string{"a", "b"}
	for _, p := range pkgs {
		write(p+"/"+p+".go", "package "+p+"\n\nfunc F() int { return 1 }\n")
		// a variable of a test file: only the run testing the package
		// rewrites it
		write(p+"/"+p+"_test.go", fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

var limit = 3

func getLimit() int { return limit }

func TestMockVar(t *testing.T) {
	mock.Patch(&limit, func() int { return 9 })
	if got := getLimit(); got != 9 {
		t.Fatalf("limit not mocked: %%d", got)
	}
}
`, p))
	}
	run := func(pkg string) (string, error) {
		cmd := exec.Command(xgo, "test", "-count=1", "./"+pkg)
		cmd.Dir = modDir
		cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		return buf.String(), err
	}
	// set up once: the instrumented GOROOT and the runtime copy
	if out, err := run("a"); err != nil {
		t.Fatalf("xgo test ./a: %v\n%s", err, out)
	}
	for round := 0; round < 4; round++ {
		var wg sync.WaitGroup
		errs := make([]string, len(pkgs))
		for i, p := range pkgs {
			wg.Add(1)
			go func(i int, p string) {
				defer wg.Done()
				if out, err := run(p); err != nil {
					errs[i] = fmt.Sprintf("xgo test ./%s: %v\n%s", p, err, out)
				}
			}(i, p)
		}
		wg.Wait()
		for _, e := range errs {
			if e != "" {
				t.Fatalf("round %d: %s", round, e)
			}
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
