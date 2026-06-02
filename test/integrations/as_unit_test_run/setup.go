package as_unit_test_run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var resolvedRootDir string
var resolveErr error

func init() {
	resolvedRootDir, resolveErr = resolveRootDir()
}

func resolveRootDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot get source file location")
	}
	dir := filepath.Dir(filename)
	dir = filepath.Join(dir, "..", "..", "..")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	if !strings.HasPrefix(string(content), "module github.com/xhd2015/xgo") {
		return "", fmt.Errorf("go.mod does not start with 'module github.com/xhd2015/xgo', wrong dir: %s", dir)
	}
	return dir, nil
}

func ResolveRootDir() (string, error) {
	if resolveErr != nil {
		return "", resolveErr
	}
	return resolvedRootDir, nil
}

func RunCommandInResolvedRootDir(name string, args ...string) error {
	root, err := ResolveRootDir()
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func goMinorVersion() (string, error) {
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected version format: %s", runtime.Version())
	}
	return parts[0] + "." + parts[1], nil
}
