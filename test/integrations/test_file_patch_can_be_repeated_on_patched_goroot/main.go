package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/xgo/test/integrations/internal"
)

func main() {
	fs := flag.NewFlagSet("test-repeat-patch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	goVersionFlag := fs.String("go-version", "", "go version to test (e.g. 1.25)")
	gorootFlag := fs.String("goroot", "", "path to GOROOT (downloads if not set)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: go run ./test/integrations/test-file-patch-can-be-repeated-on-patched-goroot --go-version 1.25 [--goroot /path/to/goroot]\n\n")
		fmt.Fprintf(fs.Output(), "Verifies that applying file-based patches twice on the same GOROOT\n")
		fmt.Fprintf(fs.Output(), "produces the same result as applying them once (idempotency).\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}
	if *goVersionFlag == "" {
		fs.Usage()
		os.Exit(1)
	}

	rootDir := internal.FindRepoRoot()
	versionName := internal.NormalizeGoVersion(*goVersionFlag)
	internal.Logf("xgo repo: %s", rootDir)
	internal.Logf("go version: %s", versionName)

	goroot, goVersion := internal.EnsureGoroot(rootDir, *goVersionFlag, *gorootFlag)
	internal.Logf("goroot: %s", goroot)

	workGoroot, cleanup := internal.SpawnWorktree(goroot, false)
	defer cleanup()

	initialCommit, err := internal.Output(workGoroot, "git", "rev-parse", "HEAD")
	if err != nil {
		internal.Fatalf("get initial commit: %v", err)
	}
	_ = initialCommit

	internal.Logf("applying patches (pass 1)...")
	if err := internal.ApplyFileBased(rootDir, workGoroot, goVersion); err != nil {
		internal.Fatalf("pass 1: %v", err)
	}

	if err := internal.RunLogged(workGoroot, nil, "git", "add", "-A", "src/"); err != nil {
		internal.Fatalf("stage pass 1: %v", err)
	}
	if err := internal.RunLogged(workGoroot, nil, "git", "commit", "--no-verify", "-m", "pass 1 [xgo]"); err != nil {
		internal.Fatalf("commit pass 1: %v", err)
	}

	commit1, err := internal.Output(workGoroot, "git", "rev-parse", "HEAD")
	if err != nil {
		internal.Fatalf("get pass 1 commit: %v", err)
	}
	commit1 = strings.TrimSpace(commit1)
	internal.Logf("pass 1 commit: %s", commit1)

	internal.Logf("applying patches (pass 2)...")
	if err := internal.ApplyFileBased(rootDir, workGoroot, goVersion); err != nil {
		internal.Fatalf("pass 2: %v", err)
	}

	internal.Logf("comparing pass 1 vs pass 2 via git diff...")
	diffOut, diffErr := internal.OutputBytes(workGoroot, "git", "diff", "HEAD", "--", "src/")
	diff := strings.TrimSpace(string(diffOut))
	if diffErr != nil && diff == "" {
		internal.Fatalf("git diff: %v", diffErr)
	}
	if diff != "" {
		cleanup = func() {} // keep worktree for inspection
		internal.Fatalf("MISMATCH — patching is not idempotent:\n%s", diff)
	}

	internal.Logf("PASS: double-patching produces identical output")
}
