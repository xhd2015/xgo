//go:build go1.24

package as_unit_test_run

import (
	"runtime"
	"strings"
	"testing"
)

func TestFilePatchCanBeRepeatedOnPatchedGoroot(t *testing.T) {
	goVersion := goMinorVersion(t)

	err := RunCommandInResolvedRootDir("go", "run", "./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot", "--go-version", goVersion)
	if err != nil {
		t.Fatal(err)
	}
}

func goMinorVersion(t *testing.T) string {
	t.Helper()
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		t.Skipf("unexpected version format: %s", runtime.Version())
	}
	return parts[0] + "." + parts[1]
}
