package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/test/integrations/internal"
)

func main() {
	fs := flag.NewFlagSet("test-compare-patch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	goVersionFlag := fs.String("go-version", "", "go version to test (1.24 or 1.25)")
	gorootFlag := fs.String("goroot", "", "path to GOROOT (downloads if not set)")
	keepWorktrees := fs.Bool("keep-worktrees", false, "always keep worktrees and temp dirs, even on success")
	includeRebuildCompilerAndGo := fs.Bool("include-rebuild-compiler-and-go", false, "include compiler and go rebuilds (skipped by default)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: go run ./test/integrations/test-file-patch-generated-same-diffs-as-programmatic-patch --go-version 1.24 [--goroot /path/to/goroot]\n\n")
		fmt.Fprintf(fs.Output(), "Compares file-based vs programmatic patching via xgo CLI.\n")
		fmt.Fprintf(fs.Output(), "Runs `xgo setup --use-file-patches=true` and `xgo setup --use-file-patches=false`\n")
		fmt.Fprintf(fs.Output(), "with separate --xgo-home dirs, then diffs the resulting GOROOTs.\n")
		fmt.Fprintf(fs.Output(), "Auto-keeps worktrees if a diff is found. Use --keep-worktrees to always keep.\n\n")
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

	goroot, _ := internal.EnsureGoroot(rootDir, *goVersionFlag, *gorootFlag)
	internal.Logf("goroot: %s", goroot)

	var cleanups []func()

	gorootFB, cleanupFB := internal.SpawnWorktree(goroot, false)
	cleanups = append(cleanups, cleanupFB)
	internal.Logf("file-based worktree: %s", gorootFB)

	gorootProg, cleanupProg := internal.SpawnWorktree(goroot, false)
	cleanups = append(cleanups, cleanupProg)
	internal.Logf("programmatic worktree: %s", gorootProg)

	xgoHomeFB, err := os.MkdirTemp("", "xgo-test-fb-*")
	if err != nil {
		internal.Fatalf("create temp dir: %v", err)
	}
	cleanups = append(cleanups, func() { os.RemoveAll(xgoHomeFB) })
	xgoHomeProg, err := os.MkdirTemp("", "xgo-test-prog-*")
	if err != nil {
		internal.Fatalf("create temp dir: %v", err)
	}
	cleanups = append(cleanups, func() { os.RemoveAll(xgoHomeProg) })

	skipRebuild := !*includeRebuildCompilerAndGo

	internal.Logf("running xgo setup --patch-goroot-in-place --use-file-patches=true...")
	gorootFB, err = internal.RunXgoSetup(xgoHomeFB, true, gorootFB, true, skipRebuild)
	if err != nil {
		runCleanups(cleanups)
		internal.Fatalf("file-based setup: %v", err)
	}
	internal.Logf("file-based goroot: %s", gorootFB)

	internal.Logf("running xgo setup --patch-goroot-in-place --use-file-patches=false...")
	gorootProg, err = internal.RunXgoSetup(xgoHomeProg, false, gorootProg, true, skipRebuild)
	if err != nil {
		runCleanups(cleanups)
		internal.Fatalf("programmatic setup: %v", err)
	}
	internal.Logf("programmatic goroot: %s", gorootProg)

	internal.Logf("creating unified comparison worktree...")
	wt, commitProg, commitFB, cleanup, err := internal.CreateUnifiedWorktree(goroot, gorootProg, gorootFB)
	if err != nil {
		runCleanups(cleanups)
		internal.Fatalf("create unified worktree: %v", err)
	}
	cleanups = append(cleanups, cleanup)

	progShort := commitProg
	if len(progShort) > 8 {
		progShort = progShort[:8]
	}
	fbShort := commitFB
	if len(fbShort) > 8 {
		fbShort = fbShort[:8]
	}
	namedDir := filepath.Join(filepath.Dir(wt), fmt.Sprintf("xgo-unified-prog-%s-vs-fb-%s", progShort, fbShort))
	if err := internal.RunLogged(goroot, nil, "git", "worktree", "move", wt, namedDir); err == nil {
		oldCleanup := cleanup
		cleanups[len(cleanups)-1] = func() {
			internal.RunLogged(goroot, nil, "git", "worktree", "remove", "--force", namedDir)
			os.RemoveAll(namedDir)
		}
		wt = namedDir
		_ = oldCleanup
	}
	internal.Logf("")
	internal.Logf("=== Unified comparison worktree ===")
	internal.Logf("Path:    %s", wt)
	internal.Logf("commit1 (programmatic): %s", commitProg)
	internal.Logf("commit2 (file-based):   %s", commitFB)
	internal.Logf("")

	internal.Logf("comparing results via git diff...")
	diff, err := internal.Output(wt, "git", "diff", commitProg+".."+commitFB, "--", "src/")
	if err != nil {
		runCleanups(cleanups)
		internal.Fatalf("git diff failed: %v", err)
	}
	diff = strings.TrimSpace(diff)
	if diff != "" {
		internal.Logf("auto-keeping worktrees due to diff (set --keep-worktrees to always keep)")
		internal.Logf("To inspect:")
		internal.Logf("  code %s", wt)
		internal.Logf("  # In VS Code, use Source Control to compare commit1 with commit2")
		internal.Logf("  # Or: git -C %s diff %s..%s -- src/", wt, commitProg, commitFB)
		internal.Fatalf("MISMATCH:\n%s", diff)
	}

	internal.Logf("PASS: file-based and programmatic produce identical output")
	if *keepWorktrees {
		internal.Logf("keeping worktrees (--keep-worktrees)")
	} else {
		runCleanups(cleanups)
	}
}

func runCleanups(cleanups []func()) {
	for _, c := range cleanups {
		if c != nil {
			c()
		}
	}
}
