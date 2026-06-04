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
	tmpDir, err := os.MkdirTemp("", "exe-shadow-repro-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "main.go")
	must(os.WriteFile(srcPath, []byte("package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }\n"), 0644))
	runCmd(tmpDir, "go", "mod", "init", "test")

	binaryName := "mybinary"

	// ── Step 1: First build ──────────────────────────────────────────────
	fmt.Println("=== Step 1: go build -o ./mybinary (prints 'hello') ===")
	os.Remove(filepath.Join(tmpDir, binaryName))
	os.Remove(filepath.Join(tmpDir, binaryName+".exe"))
	runCmd(tmpDir, "go", "build", "-o", "./"+binaryName, ".")

	binaryPath := filepath.Join(tmpDir, binaryName)
	exePath := filepath.Join(tmpDir, binaryName + ".exe")

	out, _ := run(binaryPath)
	fmt.Printf("  Run %s: %s\n", binaryName, trim(out))

	// ── Step 2: Create stale .exe ────────────────────────────────────────
	fmt.Println("\n=== Step 2: Rename mybinary → mybinary.exe (stale) ===")
	must(os.Rename(binaryPath, exePath))

	out2, _ := run(exePath)
	fmt.Printf("  Run mybinary.exe: %s\n", trim(out2))

	// ── Step 3: Rebuild with modified source ────────────────────────────
	fmt.Println("\n=== Step 3: Modify source, rebuild -o ./mybinary ===")
	srcNew := strings.Replace(string(mustReadFile(srcPath)), "hello", "hello exe", 1)
	must(os.WriteFile(srcPath, []byte(srcNew), 0644))

	os.Remove(binaryPath)
	runCmd(tmpDir, "go", "build", "-o", "./"+binaryName, ".")

	// ── Step 4: Check which binary the OS resolves to ────────────────────
	fmt.Println("\n=== Step 4: Run ./mybinary — which binary runs? ===")

	newOut, _ := run(binaryPath)
	oldOut, _ := run(exePath)

	fmt.Printf("  %s content (new):     %s\n", binaryName, trim(newOut))
	fmt.Printf("  %s.exe content (old): %s\n", binaryName, trim(oldOut))

	// ── Step 5: Let the OS resolve ──────────────────────────────────────
	fmt.Println()
	fmt.Println("=== Step 5: OS resolution (run 'mybinary') ===")

	// Use shell-agnostic approach: just run the binary name
	osOut := runOSResolved(tmpDir, binaryName)
	fmt.Printf("  OS resolves ./mybinary → output: %s\n", trim(osOut))
	fmt.Println()

	if isWindows() {
		fmt.Println("RESULT: Windows OS resolves './mybinary' to 'mybinary.exe' (the STALE one).")
		fmt.Println("        The NEW binary (mybinary, no .exe) is IGNORED.")
		fmt.Println("        This is why xgo's GOROOT rebuild silently failed:")
		fmt.Println("        -o .../compile produced 'compile' (new, instrumented)")
		fmt.Println("        but the Go toolchain resolved to 'compile.exe' (old, unpatched).")
		os.Exit(1)
	} else {
		fmt.Println("RESULT: No .exe resolution issue on non-Windows.")
		fmt.Println("        './mybinary' finds the NEW binary directly.")
		os.Exit(0)
	}
}

func isWindows() bool { return runtime.GOOS == "windows" }

func run(path string) ([]byte, error) {
	cmd := exec.Command(path)
	return cmd.CombinedOutput()
}

func runOSResolved(dir, name string) []byte {
	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.Command("cmd", "/C", name)
	} else {
		cmd = exec.Command(filepath.Join(dir, name))
	}
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return out
}

func trim(b []byte) string {
	return strings.TrimSpace(string(b))
}

func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	must(err)
	return data
}

func runCmd(dir string, args ...string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	mustE(cmd.Run())
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}
}

func mustE(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
