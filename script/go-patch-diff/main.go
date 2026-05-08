package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	downloadURLPrefix = "https://go.dev/dl/"
	revisionFileName  = "xgo-revision.txt"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var outDir string
	var bootstrapGoroot string
	var printDiff bool

	fs := flag.NewFlagSet("go-patch-diff", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&outDir, "out-dir", "go-patch-diff", "directory for scratch work and diff results")
	fs.StringVar(&bootstrapGoroot, "bootstrap-goroot", "", "GOROOT_BOOTSTRAP for building the downloaded Go source; defaults to `go env GOROOT`")
	fs.BoolVar(&printDiff, "print-diff", false, "print the final source diff to stdout instead of the output directory")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: go run ./script/go-patch-diff [flags] go1.24\n\n")
		fmt.Fprintf(fs.Output(), "Downloads a Go source tarball, snapshots it with Git, builds it,\n")
		fmt.Fprintf(fs.Output(), "runs xgo setup against it in place, then writes diff artifacts under\n")
		fmt.Fprintf(fs.Output(), "./go-patch-diff/<YYYY-MM-DD_HH-MM-SS>-<go-version>/ and prints that directory.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return fmt.Errorf("requires exactly one Go version")
	}

	rootDir, err := repoRoot()
	if err != nil {
		return err
	}
	outDir = absFromRoot(rootDir, outDir)

	versionName, err := normalizeGoVersion(fs.Arg(0))
	if err != nil {
		return err
	}

	if bootstrapGoroot == "" {
		bootstrapGoroot, err = output(rootDir, "go", "env", "GOROOT")
		if err != nil {
			return fmt.Errorf("get bootstrap GOROOT: %w", err)
		}
	}
	bootstrapGoroot, err = filepath.Abs(bootstrapGoroot)
	if err != nil {
		return err
	}

	workDir, err := makeRunDir(outDir, versionName)
	if err != nil {
		return err
	}
	resultsDir := filepath.Join(workDir, "results")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return err
	}

	logf("repo root: %s", rootDir)
	logf("work dir: %s", workDir)
	logf("bootstrap GOROOT: %s", bootstrapGoroot)

	tarName := versionName + ".src.tar.gz"
	tarPath := filepath.Join(workDir, tarName)
	tarURL := downloadURLPrefix + tarName

	logf("download %s", tarURL)
	if err := runLogged(workDir, nil, "curl", "-fL", "-o", tarPath, tarURL); err != nil {
		return err
	}

	logf("extract %s", tarName)
	if err := runLogged(workDir, nil, "tar", "-xzf", tarPath); err != nil {
		return err
	}

	extractedGo := filepath.Join(workDir, "go")
	goroot := filepath.Join(workDir, versionName)
	if err := os.Rename(extractedGo, goroot); err != nil {
		return fmt.Errorf("rename extracted go dir: %w", err)
	}

	logf("snapshot pristine source")
	if err := runLogged(goroot, nil, "git", "init", "-q"); err != nil {
		return err
	}
	if err := runLogged(goroot, nil, "git", "add", "."); err != nil {
		return err
	}
	if err := runLogged(goroot, nil,
		"git",
		"-c", "user.name=xgo patch diff",
		"-c", "user.email=xgo-patch-diff@example.invalid",
		"commit", "-q", "-m", "initial "+versionName+" source snapshot",
	); err != nil {
		return err
	}
	if err := ignoreKnownMakeArtifactDirs(goroot); err != nil {
		return err
	}

	logf("build Go source tree")
	if err := runLogged(filepath.Join(goroot, "src"), []string{"GOROOT_BOOTSTRAP=" + bootstrapGoroot}, "./make.bash"); err != nil {
		return err
	}
	if err := ignoreUntrackedMakeArtifacts(goroot); err != nil {
		return err
	}

	goVersion, err := output(goroot, filepath.Join(goroot, "bin", "go"), "version")
	if err != nil {
		return fmt.Errorf("verify built go: %w", err)
	}
	logf("built %s", goVersion)

	// Make xgo treat this source tree as an already-instrumented GOROOT so
	// setup patches it in place instead of copying it to ~/.xgo/go-instrument.
	if err := os.WriteFile(filepath.Join(workDir, revisionFileName), []byte("initial-snapshot-for-in-place-diff\n"), 0644); err != nil {
		return err
	}

	logf("run xgo setup in place")
	if err := runLogged(rootDir, nil,
		"go", "run", "-tags=dev", "./cmd/xgo",
		"setup",
		"--with-goroot", goroot,
		"--reset-instrument",
		"--log-debug=stdout",
		"-v",
	); err != nil {
		return err
	}

	logf("collect diffs")
	if err := writeCommandOutput(resultsDir, "tracked.diff", goroot, "git", "diff"); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "tracked.stat.txt", goroot, "git", "diff", "--stat"); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "tracked.name-status.txt", goroot, "git", "diff", "--name-status"); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "status-before-intent-to-add.txt", goroot, "git", "status", "--short"); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "untracked.txt", goroot, "git", "ls-files", "--others", "--exclude-standard"); err != nil {
		return err
	}

	xgoNewFiles, err := xgoUntrackedFiles(goroot)
	if err != nil {
		return err
	}
	if len(xgoNewFiles) > 0 {
		addArgs := append([]string{"add", "-N", "--"}, xgoNewFiles...)
		if err := runLogged(goroot, nil, "git", addArgs...); err != nil {
			return err
		}
	}

	finalDiff, err := outputBytes(goroot, "git", "diff")
	if err != nil {
		return err
	}
	if err := writeFile(resultsDir, "source-with-new-files.diff", finalDiff); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "source-with-new-files.stat.txt", goroot, "git", "diff", "--stat"); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "source-with-new-files.name-status.txt", goroot, "git", "diff", "--name-status"); err != nil {
		return err
	}
	if err := writeCommandOutput(resultsDir, "status-after-intent-to-add.txt", goroot, "git", "status", "--short"); err != nil {
		return err
	}

	logf("final diff: %s", filepath.Join(resultsDir, "source-with-new-files.diff"))
	logf("diff stat: %s", filepath.Join(resultsDir, "source-with-new-files.stat.txt"))
	logf("patched tree: %s", goroot)

	if printDiff {
		_, err = os.Stdout.Write(finalDiff)
		return err
	}
	fmt.Fprintln(os.Stdout, workDir)
	return nil
}

func repoRoot() (string, error) {
	root, err := output("", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("find repo root: %w", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "xgo", "main.go")); err != nil {
		return "", fmt.Errorf("repo root %s does not look like xgo: %w", root, err)
	}
	return root, nil
}

func absFromRoot(rootDir string, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(rootDir, path)
}

var minorVersionOnly = regexp.MustCompile(`^\d+\.\d+$`)

func normalizeGoVersion(version string) (string, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return "", fmt.Errorf("empty Go version")
	}
	version = strings.TrimPrefix(version, "go")
	if minorVersionOnly.MatchString(version) {
		version += ".0"
	}
	if !strings.HasPrefix(version, "1.") {
		return "", fmt.Errorf("unsupported Go version %q: expected go1.x or 1.x", version)
	}
	return "go" + version, nil
}

func makeRunDir(outDir string, versionName string) (string, error) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}
	base := time.Now().Format("2006-01-02_15-04-05") + "-" + versionName
	for i := 0; ; i++ {
		name := base
		if i > 0 {
			name = fmt.Sprintf("%s-%02d", base, i)
		}
		dir := filepath.Join(outDir, name)
		err := os.Mkdir(dir, 0755)
		if err == nil {
			return dir, nil
		}
		if !os.IsExist(err) {
			return "", err
		}
	}
}

func ignoreKnownMakeArtifactDirs(goroot string) error {
	return appendInfoExcludePatterns(goroot, "# xgo go-patch-diff: known directories generated by make.bash", []string{
		"/bin/",
		"/pkg/",
	})
}

func ignoreUntrackedMakeArtifacts(goroot string) error {
	untracked, err := gitUntrackedFiles(goroot)
	if err != nil {
		return err
	}
	if len(untracked) == 0 {
		return nil
	}

	patterns := make([]string, 0, len(untracked))
	for _, file := range untracked {
		pattern := "/" + filepath.ToSlash(file)
		if info, statErr := os.Stat(filepath.Join(goroot, file)); statErr == nil && info.IsDir() {
			pattern += "/"
		}
		patterns = append(patterns, pattern)
	}
	return appendInfoExcludePatterns(goroot, "# xgo go-patch-diff: auto-detected artifacts generated by make.bash", patterns)
}

func appendInfoExcludePatterns(goroot string, marker string, patterns []string) error {
	excludeFile := filepath.Join(goroot, ".git", "info", "exclude")
	existingBytes, err := os.ReadFile(excludeFile)
	if err != nil {
		return err
	}

	existing := string(existingBytes)
	existingPatterns := make(map[string]bool)
	for _, line := range strings.Split(existing, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			existingPatterns[line] = true
		}
	}

	var filtered []string
	for _, pattern := range patterns {
		if pattern == "" || existingPatterns[pattern] {
			continue
		}
		filtered = append(filtered, pattern)
	}
	if len(filtered) == 0 {
		return nil
	}

	logf("update %s: add %d ignore pattern(s): %s", excludeFile, len(filtered), summarizePatterns(filtered))

	var b strings.Builder
	b.WriteString(existing)
	if len(existingBytes) > 0 && !strings.HasSuffix(existing, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	b.WriteString(marker)
	b.WriteByte('\n')
	for _, pattern := range filtered {
		b.WriteString(pattern)
		b.WriteByte('\n')
	}

	return os.WriteFile(excludeFile, []byte(b.String()), 0644)
}

func summarizePatterns(patterns []string) string {
	const maxPatterns = 8
	if len(patterns) <= maxPatterns {
		return strings.Join(patterns, ", ")
	}
	return strings.Join(patterns[:maxPatterns], ", ") + fmt.Sprintf(", ... +%d more", len(patterns)-maxPatterns)
}

func xgoUntrackedFiles(goroot string) ([]string, error) {
	untracked, err := gitUntrackedFiles(goroot)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range untracked {
		slashPath := filepath.ToSlash(line)
		base := filepath.Base(slashPath)
		if slashPath == ".gitignore" ||
			strings.Contains(slashPath, "xgo_rewrite_internal/") ||
			strings.Contains(base, "xgo") {
			files = append(files, line)
		}
	}
	return files, nil
}

func gitUntrackedFiles(goroot string) ([]string, error) {
	out, err := output(goroot, "git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, line)
	}
	return files, nil
}

func writeCommandOutput(resultsDir string, name string, dir string, cmdName string, args ...string) error {
	out, err := outputBytes(dir, cmdName, args...)
	if err != nil {
		return err
	}
	return writeFile(resultsDir, name, out)
}

func writeFile(dir string, name string, data []byte) error {
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}

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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, commandError(cmdName, args, stderr.String(), err)
	}
	return out, nil
}

func runLogged(dir string, env []string, cmdName string, args ...string) error {
	logf("+ %s", commandString(cmdName, args))
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return commandError(cmdName, args, "", err)
	}
	return nil
}

func commandError(cmdName string, args []string, stderr string, err error) error {
	msg := strings.TrimSpace(stderr)
	if msg != "" {
		return fmt.Errorf("%s: %w\n%s", commandString(cmdName, args), err, msg)
	}
	return fmt.Errorf("%s: %w", commandString(cmdName, args), err)
}

func commandString(cmdName string, args []string) string {
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
	if strings.IndexFunc(s, func(r rune) bool {
		return !(r == '/' || r == '.' || r == '-' || r == '_' || r == '=' || r == ':' ||
			('0' <= r && r <= '9') ||
			('A' <= r && r <= 'Z') ||
			('a' <= r && r <= 'z'))
	}) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
