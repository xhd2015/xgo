package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/instrument/instrument_compiler"
	"github.com/xhd2015/xgo/instrument/instrument_go"
	"github.com/xhd2015/xgo/instrument/instrument_runtime"
	patches "github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

const (
	downloadURLPrefix = "https://go.dev/dl/"
)

func main() {
	fs := flag.NewFlagSet("file-based-patch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	goVersionFlag := fs.String("go-version", "", "go version to test (1.24 or 1.25)")
	gorootFlag := fs.String("goroot", "", "path to GOROOT (downloads if not set)")
	keepTemp := fs.Bool("keep-temp", false, "keep temp directories for inspection")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: go run ./test/file-based-patch --go-version 1.24 [--goroot /path/to/groot]\n\n")
		fmt.Fprintf(fs.Output(), "Compares programmatic vs file-based patching for a Go version.\n")
		fmt.Fprintf(fs.Output(), "If --goroot is not set, downloads and builds the Go source tarball.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}
	if *goVersionFlag == "" {
		fs.Usage()
		os.Exit(1)
	}

	rootDir := findRepoRoot()
	versionName := normalizeGoVersion(*goVersionFlag)
	logf("xgo repo: %s", rootDir)
	logf("go version: %s", versionName)

	goroot, err := resolveGoroot(rootDir, versionName, *gorootFlag)
	if err != nil {
		fatalf("%v", err)
	}
	logf("goroot: %s", goroot)

	goVersion, err := goinfo.GetGorootVersion(goroot)
	if err != nil {
		fatalf("get goroot version: %v", err)
	}

	logf("creating temp dirs...")
	dirA, err := os.MkdirTemp("", "xgo-test-prog-*")
	if err != nil {
		fatalf("create temp dir: %v", err)
	}
	dirB, err := os.MkdirTemp("", "xgo-test-file-*")
	if err != nil {
		fatalf("create temp dir: %v", err)
	}
	if !*keepTemp {
		defer os.RemoveAll(dirA)
		defer os.RemoveAll(dirB)
	}

	gorootBase := filepath.Base(goroot)
	gorootA := filepath.Join(dirA, gorootBase)
	gorootB := filepath.Join(dirB, gorootBase)

	logf("sync goroot to temp dirs...")
	if err := cpGoroot(goroot, gorootA); err != nil {
		fatalf("sync goroot A: %v", err)
	}
	if err := cpGoroot(goroot, gorootB); err != nil {
		fatalf("sync goroot B: %v", err)
	}

	logf("applying programmatic patches...")
	if err := applyProgrammatic(goroot, gorootA, rootDir, goVersion); err != nil {
		fatalf("programmatic: %v", err)
	}

	logf("applying file-based patches...")
	if err := applyFileBased(rootDir, gorootB, goVersion); err != nil {
		fatalf("file-based: %v", err)
	}

	logf("comparing results...")
	diff := compareDirs(gorootA, gorootB)
	if diff != "" {
		fatalf("MISMATCH:\n%s", diff)
	}

	logf("PASS: programmatic and file-based produce identical output")
}

func findRepoRoot() string {
	root, err := output("", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		fatalf("find repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "xgo", "main.go")); err != nil {
		fatalf("not xgo repo root: %s", root)
	}
	return root
}

func normalizeGoVersion(version string) string {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "go")
	if !strings.Contains(v, ".") {
		fatalf("invalid version: %q", version)
	}
	parts := strings.Split(v, ".")
	if len(parts) == 2 {
		v += ".0"
	}
	return "go" + v
}

func resolveGoroot(rootDir string, versionName string, gorootFlag string) (string, error) {
	if gorootFlag != "" {
		return gorootFlag, nil
	}

	cacheDir := filepath.Join(rootDir, "test", "file-based-patch", "cache", versionName)
	if _, err := os.Stat(filepath.Join(cacheDir, "bin", "go")); err == nil {
		logf("using cached goroot: %s", cacheDir)
		return cacheDir, nil
	}

	logf("downloading %s...", versionName)
	tarName := versionName + ".src.tar.gz"
	tarPath := filepath.Join(cacheDir, tarName)
	os.MkdirAll(cacheDir, 0755)

	tarURL := downloadURLPrefix + tarName
	if err := runLogged(cacheDir, nil, "curl", "-fL", "-o", tarPath, tarURL); err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	logf("extracting...")
	if err := runLogged(cacheDir, nil, "tar", "-xzf", tarPath); err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}

	extractedDir := filepath.Join(cacheDir, "go")
	srcDir := versionName
	if err := os.Rename(extractedDir, srcDir); err != nil {
		return "", fmt.Errorf("rename extracted dir: %w", err)
	}
	goroot := filepath.Join(cacheDir, srcDir)

	logf("building go toolchain...")
	bootstrapGoroot := os.Getenv("GOROOT_BOOTSTRAP")
	if bootstrapGoroot == "" {
		out, err := output(rootDir, "go", "env", "GOROOT")
		if err != nil {
			return "", fmt.Errorf("get bootstrap GOROOT: %w (set GOROOT_BOOTSTRAP env)", err)
		}
		bootstrapGoroot = out
	}
	if err := runLogged(filepath.Join(goroot, "src"), []string{"GOROOT_BOOTSTRAP=" + bootstrapGoroot}, "./make.bash"); err != nil {
		return "", fmt.Errorf("make.bash: %w", err)
	}

	return goroot, nil
}

func cpGoroot(src, dst string) error {
	return exec.Command("cp", "-R", src, dst).Run()
}

func applyProgrammatic(origGoroot, goroot, xgoSrc string, goVersion *goinfo.GoVersion) error {
	if err := instrument_go.BuildInstrument(goroot, goVersion, true); err != nil {
		return fmt.Errorf("BuildInstrument: %w", err)
	}
	if err := instrument_go.InstrumentGoToolCover(goroot, goVersion); err != nil {
		return fmt.Errorf("InstrumentGoToolCover: %w", err)
	}
	trapPath := filepath.Join(xgoSrc, "cmd", "xgo", "asset", "runtime_gen", "internal", "runtime", "xgo_trap_template.go")
	trapBytes, err := os.ReadFile(trapPath)
	if err != nil {
		return fmt.Errorf("read trap template: %w", err)
	}
	if err := instrument_runtime.InstrumentRuntime(goroot, goVersion, string(trapBytes), instrument_runtime.InstrumentRuntimeOptions{
		Mode: instrument_runtime.InstrumentMode_ForceAndIgnoreMark,
	}); err != nil {
		return fmt.Errorf("InstrumentRuntime: %w", err)
	}
	if err := instrument_compiler.BuildInstrument(origGoroot, goroot, goVersion, xgoSrc, true, true); err != nil {
		return fmt.Errorf("BuildInstrument: %w", err)
	}
	if _, err := build.BuildGoToolCompileDebugBinary(goroot); err != nil {
		return fmt.Errorf("BuildGoToolCompileDebugBinary: %w", err)
	}
	return nil
}

func applyFileBased(xgoSrc, goroot string, goVersion *goinfo.GoVersion) error {
	srcDir := filepath.Join(xgoSrc, "patches", fmt.Sprintf("go%d.%d", goVersion.Major, goVersion.Minor))
	tmpPatchDir, err := os.MkdirTemp("", "xgo-patch-test-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpPatchDir)
	runLogged("", nil, "cp", "-R", srcDir+"/", tmpPatchDir+"/")

	// Override config: empty generate (skip rebuilds for diff comparison)
	config := map[string]interface{}{
		"version":  fmt.Sprintf("go%d.%d", goVersion.Major, goVersion.Minor),
		"copy":     jsonCopyEntries(srcDir),
		"generate": []interface{}{},
	}
	configBytes, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(tmpPatchDir, "__config__.json"), configBytes, 0644)

	extraEnv := map[string]string{
		"XGO_SRC":     xgoSrc,
		"GOROOT":      goroot,
		"ORIG_GOROOT": goroot,
		"GO_VERSION":  fmt.Sprintf("go%d.%d.%d", goVersion.Major, goVersion.Minor, goVersion.Patch),
		"GOOS":        runtime.GOOS,
		"GOARCH":      runtime.GOARCH,
	}
	return patches.ApplyPatches(tmpPatchDir, goroot, xgoSrc, extraEnv)
}

func jsonCopyEntries(srcDir string) interface{} {
	// Read the original config to get the copy entries
	configPath := filepath.Join(srcDir, "__config__.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return []interface{}{}
	}
	var cfg struct {
		Copy json.RawMessage `json:"copy"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return []interface{}{}
	}
	if len(cfg.Copy) > 0 && cfg.Copy[0] == '[' {
		var entries []map[string]interface{}
		json.Unmarshal(cfg.Copy, &entries)
		return entries
	}
	return []interface{}{}
}

func compareDirs(dirA, dirB string) string {
	// Strip inline markers from both dirs before comparing.
	// Inline markers differ in placement between programmatic and file-based approaches
	// but the underlying code changes should be equivalent.
	stripA := stripMarkers(dirA)
	stripB := stripMarkers(dirB)
	defer os.RemoveAll(stripA)
	defer os.RemoveAll(stripB)

	base := filepath.Base(dirA)
	cmd := exec.Command("diff", "-rq",
		"--exclude=.DS_Store",
		"--exclude=bin",
		"--exclude=pkg",
		"--exclude=.git",
		filepath.Join(stripA, base, "src"),
		filepath.Join(stripB, base, "src"))
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

func stripMarkers(dir string) string {
	dst, _ := os.MkdirTemp("", "xgo-compare-*")
	exec.Command("cp", "-R", dir, dst).Run()

	walkRoot := filepath.Join(dst, filepath.Base(dir))
	// Strip /*<...>*/ markers from Go files
	filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)
		// Remove all /*<anything>*/ markers (inline idempotency markers)
		for {
			start := strings.Index(text, "/*<")
			if start < 0 {
				break
			}
			end := strings.Index(text[start:], "*/")
			if end < 0 {
				break
			}
			end += start + 2
			text = text[:start] + text[end:]
		}
		os.WriteFile(path, []byte(text), 0644)
		return nil
	})
	return dst
}

// === Helpers ===

func output(dir string, cmdName string, args ...string) (string, error) {
	out, err := outputBytes(dir, cmdName, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func outputBytes(dir string, cmdName string, args ...string) ([]byte, error) {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%s: %w\n%s", cmd, err, stderr.String())
	}
	return out, nil
}

func runLogged(dir string, env []string, cmdName string, args ...string) error {
	logf("+ %s", shellQuoteArgs(cmdName, args))
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %w", cmd, err)
	}
	return nil
}

func shellQuoteArgs(cmdName string, args []string) string {
	parts := append([]string{cmdName}, args...)
	for i, p := range parts {
		parts[i] = shellQuote(p)
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if !(r == '/' || r == '.' || r == '-' || r == '_' || r == '=' || r == ':' ||
			('0' <= r && r <= '9') ||
			('A' <= r && r <= 'Z') ||
			('a' <= r && r <= 'z')) {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}

func logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[test] "+format+"\n", args...)
}

var fatalf = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}


