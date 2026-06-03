package main

import (
	"bytes"
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

	// Run with \n separator
	nlContent, nlDir := runXgoTest(repoRoot, xgoBin, proofSrc, proofTestSrc, nil, "\\n")
	// Run with ; separator
	semiContent, semiDir := runXgoTest(repoRoot, xgoBin, proofSrc, proofTestSrc,
		[]string{"XGO_DEBUG_USE_SEMICOLON=true"}, ";")

	// Diff
	fmt.Println("\n=== DIFF (\\n vs ;) ===")
	fmt.Println(diffStrings(nlContent, semiContent))

	fmt.Println("\n=== OVERLAY FILES ===")
	fmt.Printf("\\n version: %s\n", filepath.Join(nlDir, ".xgo", "gen", "overlay", "PROJECT", "proof.go"))
	fmt.Printf(";  version: %s\n", filepath.Join(semiDir, ".xgo", "gen", "overlay", "PROJECT", "proof.go"))

	fmt.Println("=== DONE ===")
}

func runXgoTest(repoRoot, xgoBin, proofSrc, proofTestSrc string, env []string, label string) (string, string) {
	tmpDir, err := os.MkdirTemp("", "xgo-cr-*")
	if err != nil {
		fatalf("[%s] create temp dir: %v", label, err)
	}
	// Keep dir for inspection
	fmt.Printf("[%s] test dir: %s\n", label, tmpDir)

	goModContent := fmt.Sprintf("module test\n\ngo 1.21\n\nrequire github.com/xhd2015/xgo/runtime v0.0.0\n\nreplace github.com/xhd2015/xgo/runtime => %s/runtime\n", repoRoot)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "proof.go"), []byte(proofSrc), 0644)
	os.WriteFile(filepath.Join(tmpDir, "proof_test.go"), []byte(proofTestSrc), 0644)

	if out, err := runCmdEnv(tmpDir, nil, "go", "mod", "tidy"); err != nil {
		fatalf("[%s] go mod tidy: %v\n%s", label, err, out)
	}

	fmt.Printf("[%s] Running xgo test...\n", label)
	out, err := runCmdEnv(tmpDir, env, xgoBin, "test", "--trap-all", "-a", "-v", ".")
	if err != nil {
		fmt.Printf("[%s] xgo test output:\n%s\n", label, out)
	}
	// Don't fatal on error — we want to collect the overlay even if build fails

	// Read instrumented proof.go from overlay
	overlayProofPath := filepath.Join(tmpDir, ".xgo", "gen", "overlay", "PROJECT", "proof.go")
	content, err := os.ReadFile(overlayProofPath)
	if err != nil {
		fmt.Printf("[%s] WARNING: no overlay proof.go: %v\n", label, err)
		return fmt.Sprintf("(no overlay: %v)", err), tmpDir
	}

	crCount := strings.Count(string(content), "\r")
	fmt.Printf("[%s] instrumented proof.go, %d \\r bytes, %d lines\n", label, crCount, strings.Count(string(content), "\n"))
	return string(content), tmpDir
}

func diffStrings(a, b string) string {
	fileA, _ := os.CreateTemp("", "xgo-diff-a-*")
	fileB, _ := os.CreateTemp("", "xgo-diff-b-*")
	defer os.Remove(fileA.Name())
	defer os.Remove(fileB.Name())
	os.WriteFile(fileA.Name(), []byte(a), 0644)
	os.WriteFile(fileB.Name(), []byte(b), 0644)

	cmd := exec.Command("diff", "-u", fileA.Name(), fileB.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// diff returns non-zero when files differ
		if stdout.Len() > 0 {
			return stdout.String()
		}
		return stderr.String()
	}
	return "(no differences)"
}

func ensureXgo(repoRoot string) (binPath string, cleanup func()) {
	// Always build fresh to pick up latest instrument/ changes
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

func runCmdEnv(dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	var stderr, stdout bytes.Buffer
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
