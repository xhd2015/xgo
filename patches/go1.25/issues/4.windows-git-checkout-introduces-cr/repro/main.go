package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Source file with LF-only line endings, containing a raw string literal.
// On Windows with core.autocrlf=true, git-clone will convert \n to \r\n,
// injecting \r bytes INTO the raw string literal.
const srcPayload = "package p\n\nconst help = `\nUsage:\n  -func match\n`\n\nfunc Hello() {}\n"

func main() {
	// ── Step 1: show git autocrlf config ──────────────────────────────────
	fmt.Println("=== Git autocrlf config ===")
	showConfig()

	// ── Step 2: create temp repo with LF-only source ──────────────────────
	tmpDir, err := os.MkdirTemp("", "git-crlf-repro-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	repoDir := filepath.Join(tmpDir, "repo")
	must(os.MkdirAll(repoDir, 0755))

	srcPath := filepath.Join(repoDir, "main.go")
	must(os.WriteFile(srcPath, []byte(srcPayload), 0644))

	fmt.Println("\n=== Before commit (source file) ===")
	checkFile("source", srcPath)

	// ── Step 3: git init + commit with autocrlf=false ─────────────────────
	// This ensures the committed blob is LF-only regardless of platform.
	git := func(args ...string) { runGit(repoDir, args...) }
	git("-c", "core.autocrlf=false", "init")
	git("-c", "core.autocrlf=false", "config", "user.email", "test@test")
	git("-c", "core.autocrlf=false", "config", "user.name", "test")
	git("-c", "core.autocrlf=false", "add", "main.go")
	git("-c", "core.autocrlf=false", "commit", "-m", "initial")

	fmt.Println("\n=== After commit (working tree) ===")
	checkFile("committed", srcPath)

	// ── Step 4: clone with platform-default autocrlf (no -c override) ────
	cloneDir := filepath.Join(tmpDir, "clone")
	must(os.MkdirAll(cloneDir, 0755))
	runGit(tmpDir, "clone", repoDir, cloneDir)

	clonePath := filepath.Join(cloneDir, "main.go")

	fmt.Println("\n=== After clone (platform-default autocrlf) ===")
	foundCR := checkFile("cloned", clonePath)

	// ── Step 5: verdict ───────────────────────────────────────────────────
	fmt.Println()
	if foundCR {
		fmt.Println("RESULT: \\r FOUND — git autocrlf altered the source file.")
		fmt.Println("        Raw string literals now contain \\r bytes.")
		fmt.Println("        This is the ROOT CAUSE of xgo's position-calculation bugs.")
		fmt.Println("        See: patches/go1.25/issues/windows-raw-string-syntax-rewrite-issue/")
		os.Exit(1)
	}
	fmt.Println("RESULT: No \\r found — file is pure LF. Safe.")
}

func showConfig() {
	cmd := exec.Command("git", "config", "--list", "--show-origin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	_ = cmd.Run()
}

func checkFile(label, path string) bool {
	data, err := os.ReadFile(path)
	must(err)

	crCount := bytes.Count(data, []byte("\r"))
	lfCount := bytes.Count(data, []byte("\n"))

	fmt.Printf("  %s: %d bytes, \\r=%d, \\n=%d\n", label, len(data), crCount, lfCount)

	if crCount > 0 {
		fmt.Println("  Hex dump (first 200 bytes):")
		hexDump(data, 200)
		return true
	}
	return false
}

func hexDump(data []byte, maxLen int) {
	n := len(data)
	if n > maxLen {
		n = maxLen
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		b := data[i]
		if b == '\r' {
			sb.WriteString("[0d]")
		} else if b == '\n' {
			sb.WriteString("[0a]")
		} else if b >= 0x20 && b < 0x7f {
			sb.WriteByte(b)
		} else {
			sb.WriteString(fmt.Sprintf("[%02x]", b))
		}
	}
	fmt.Printf("  %s\n", sb.String())
	fmt.Println("  Legend: [0d]=\\r  [0a]=\\n")
}

func runGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}
}
