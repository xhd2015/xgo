package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xhd2015/xgo/test/integrations/internal"
)

// Minimal repro for: src.NoXPos_xgo_get undefined (or GOTOOLCHAIN version mismatch)
// during xgo setup --reset-instrument when xgo is in PATH.
//
// The bug: apply.go runs ${INSTRUMENT_GOROOT}/bin/go build ... cmd/compile
// without GO_BYPASS_XGO=true. The instrumented go binary's xgoPrecheck hook
// finds xgo in PATH, delegates back to xgo, which tries to re-instrument
// toolchain packages. This fails because variable trapping generates getter
// references that aren't available for stdlib packages.
//
// Reproduction requires passing the INSTRUMENTED goroot to --reset-instrument
// (so instrumented=true and the dir is reused, not deleted).
//
// Usage:
//   go run ./test/integrations/xgo-with-setup-repro/ --goroot go-release/go1.25.10

func main() {
	goVersionFlag := flag.String("go-version", "", "go version to download if --goroot not set (e.g. 1.25)")
	gorootFlag := flag.String("goroot", "", "path to ORIGINAL GOROOT (downloads if not set)")
	flag.Parse()

	repoRoot := internal.FindRepoRoot()
	internal.Logf("xgo repo: %s", repoRoot)

	if *goVersionFlag == "" && *gorootFlag == "" {
		fmt.Fprintf(os.Stderr, "error: --go-version or --goroot is required\n")
		os.Exit(1)
	}

	goroot, _ := internal.EnsureGoroot(repoRoot, *goVersionFlag, *gorootFlag)
	internal.Logf("original goroot: %s", goroot)

	xgoHome, err := os.MkdirTemp("", "xgo-repro-*")
	if err != nil {
		internal.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(xgoHome)
	internal.Logf("xgo home: %s", xgoHome)

	// install xgo to PATH
	internal.Logf("installing xgo...")
	if err := internal.RunLogged(repoRoot, nil, "go", "install", "./cmd/xgo"); err != nil {
		internal.Fatalf("go install xgo: %v", err)
	}

	// first setup — creates instrumented GOROOT
	internal.Logf("=== first setup (fresh) ===")
	instrumentGoroot, err := runXgoSetup(repoRoot, xgoHome, goroot, false)
	if err != nil {
		internal.Fatalf("first setup failed: %v", err)
	}
	internal.Logf("instrument goroot: %s", instrumentGoroot)

	// second setup — pass the INSTRUMENTED goroot so instrumented=true
	// and the existing go binary (with xgoPrecheck hook) is used for rebuild.
	internal.Logf("=== second setup (reset-instrument with instrumented goroot) ===")
	instrumentGoroot2, err := runXgoSetup(repoRoot, xgoHome, instrumentGoroot, true)
	if err != nil {
		internal.Fatalf("FAIL: second setup failed (BUG PRESENT): %v", err)
	}
	internal.Logf("instrument goroot: %s", instrumentGoroot2)

	fmt.Println("PASS: double setup succeeded (bug fixed or not reproducing)")
}

func runXgoSetup(repoRoot, xgoHome, goroot string, reset bool) (string, error) {
	args := []string{
		"run",
		"./cmd/xgo", "setup",
		"--xgo-home", xgoHome,
		"--with-goroot", goroot,

	}
	if reset {
		args = append(args, "--reset-instrument")
	}

	internal.Logf("+ go %v", args)
	out, err := internal.Output(repoRoot, "go", args...)
	if err != nil {
		return "", err
	}
	return out, nil
}
