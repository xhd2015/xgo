//go:build go1.24 && !go1.26

package as_unit_test_run

import (
	"runtime"
	"strings"
	"testing"
)

func TestFilePatchGeneratedSameDiffsAsProgramaticPatch(t *testing.T) {
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		t.Skipf("unexpected version format: %s", runtime.Version())
	}
	if parts[0] != "1" || (parts[1] != "24" && parts[1] != "25") {
		t.Skipf("test only runs on go1.24 and go1.25, got %s", runtime.Version())
	}
	goVersion := parts[0] + "." + parts[1]

	err := RunCommandInResolvedRootDir("go", "run", "./test/integrations/test_file_patch_generated_same_diffs_as_programatic_patch", "--go-version", goVersion)
	if err != nil {
		t.Fatal(err)
	}
}
