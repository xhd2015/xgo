package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// \r alone inside interpreted string — VALID Go
	srcR := "package p\n\nconst s = \"hello\x0dworld\"\n"

	// \r\n inside interpreted string — INVALID (newline in string)
	srcRNL := "package p\n\nconst s = \"hello\x0d\x0aworld\"\n"

	fmt.Println("=== Test A: \\r only inside \"...\" (expected: compiles) ===")
	testCompile(srcR, true, "\\r only")

	fmt.Println("\n=== Test B: \\r\\n inside \"...\" (expected: fails) ===")
	testCompile(srcRNL, false, "\\r\\n")

	fmt.Println("\n=== Test C: \\r\\n inside `...` (expected: compiles) ===")
	srcRaw := "package p\n\nconst s = `hello\x0d\x0aworld`\n"
	testCompile(srcRaw, true, "\\r\\n in raw string")

	fmt.Println("\n=== ALL TESTS PASSED ===")
}

func testCompile(src string, expectPass bool, label string) {
	tmpDir, err := os.MkdirTemp("", "xgo-string-cr-*")
	if err != nil {
		fatalf("create temp dir: %v", err)
	}
	fmt.Printf("  dir: %s\n", tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(src), 0644)
	// Also write hex dump for inspection
	hexFile := filepath.Join(tmpDir, "main.go.hex")
	os.WriteFile(hexFile, []byte(fmt.Sprintf("%x", src)), 0644)

	cmd := exec.Command("go", "build", ".")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()

	if err != nil {
		firstLine := string(out)
		if idx := indexByte(firstLine, '\n'); idx >= 0 {
			firstLine = firstLine[:idx]
		}
		if expectPass {
			fmt.Printf("  UNEXPECTED FAIL: %s\n", firstLine)
			fatalf("expected PASS, got FAIL for %s", label)
		}
		fmt.Printf("  FAIL (expected): %s\n", firstLine)
	} else {
		if !expectPass {
			fatalf("expected FAIL, got PASS for %s", label)
		}
		fmt.Println("  PASS")
	}
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}
