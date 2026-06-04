package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	repoRoot := findRepoRoot()

	xgoBin, cleanup := ensureXgo(repoRoot)
	defer cleanup()

	proofSrc := "package proof\r\n\r\nimport _ \"github.com/xhd2015/xgo/runtime/trap\"\r\n\r\nvar help = `\r\nUsage:\r\n  -func match\r\n`\r\n\r\nfunc GetHelp() string {\r\n\treturn help\r\n}\r\n"
	proofTestSrc := "package proof\r\n\r\nimport (\r\n\t\"testing\"\r\n)\r\n\r\nfunc TestGetHelp(t *testing.T) {\r\n\tresult := GetHelp()\r\n\t_ = result\r\n}\r\n"

	if !strings.Contains(proofSrc, "\r") {
		fatalf("source does not contain \\r bytes")
	}
	fmt.Printf("proof.go has %d \\r bytes\n", strings.Count(proofSrc, "\r"))
	fmt.Printf("proof_test.go has %d \\r bytes\n", strings.Count(proofTestSrc, "\r"))

	tmpDir, err := os.MkdirTemp("", "xgo-cr-*")
	if err != nil {
		fatalf("create temp dir: %v", err)
	}
	// Keep dir for inspection
	fmt.Printf("test dir: %s\n", tmpDir)

	goModContent := fmt.Sprintf("module test\n\ngo 1.21\n\nrequire github.com/xhd2015/xgo/runtime v0.0.0\n\nreplace github.com/xhd2015/xgo/runtime => %s/runtime\n", repoRoot)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "proof.go"), []byte(proofSrc), 0644)
	os.WriteFile(filepath.Join(tmpDir, "proof_test.go"), []byte(proofTestSrc), 0644)

	if out, err := runCmd(tmpDir, "go", "mod", "tidy"); err != nil {
		fatalf("go mod tidy: %v\n%s", err, out)
	}

	fmt.Println("Running xgo test --trap-all ...")
	out, err := runCmd(tmpDir, xgoBin, "test", "--trap-all", "-a", "-v", ".")
	if err != nil {
		fmt.Printf("xgo test output:\n%s\n", out)
	}
	// Don't fatal on error — we want to collect the overlay even if build fails

	overlayPath := filepath.Join(tmpDir, ".xgo", "gen", "overlay", "PROJECT", "proof.go")
	content, err := os.ReadFile(overlayPath)
	if err != nil {
		fmt.Printf("WARNING: no overlay proof.go: %v\n", err)
	} else {
		crCount := strings.Count(string(content), "\r")
		fmt.Printf("instrumented proof.go, %d \\r bytes, %d lines\n", crCount, strings.Count(string(content), "\n"))
		// Print a snippet around the inserted trap/stub area
		idx := strings.Index(string(content), "func help_xgo_get")
		if idx >= 0 {
			start := idx - 50
			if start < 0 {
				start = 0
			}
			end := idx + 200
			if end > len(content) {
				end = len(content)
			}
			fmt.Printf("snippet around insertion:\n---\n%s\n---\n", string(content[start:end]))
		}
	}

	fmt.Printf("overlay proof.go: %s\n", overlayPath)
	fmt.Println("=== DONE ===")
}

func ensureXgo(repoRoot string) (binPath string, cleanup func()) {
	tmpDir, err := os.MkdirTemp("", "xgo-build-*")
	if err != nil {
		fatalf("create build dir: %v", err)
	}
	binName := "xgo"
	if runtime.GOOS == "windows" {
		binName = "xgo.exe"
	}
	binPath = filepath.Join(tmpDir, binName)
	fmt.Printf("Building xgo to %s...\n", binPath)
	buildCmd := exec.Command("go", "build", "-a", "-o", binPath, "./cmd/xgo")
	buildCmd.Dir = repoRoot
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdout = os.Stdout
	if err := buildCmd.Run(); err != nil {
		fatalf("build xgo: %v", err)
	}
	return binPath, func() { os.RemoveAll(tmpDir) }
}

func findRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fatalf("could not find repo root (go.mod)")
		}
		dir = parent
	}
}

func runCmd(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stderr, stdout strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	combined := stdout.String()
	if stderr.Len() > 0 {
		combined += "\n[stderr]\n" + stderr.String()
	}
	return combined, err
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}
