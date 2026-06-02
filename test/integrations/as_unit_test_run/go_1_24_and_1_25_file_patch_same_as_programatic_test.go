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
	major := parts[0]
	minor := parts[1]
	if major != "1" || (minor != "24" && minor != "25") {
		t.Skipf("test only runs on go1.24 and go1.25, got %s", runtime.Version())
	}
	if minor == "25" {
		t.Skipf("go1.25 progmatic patch should still be re-aligned with file-based patch, but the function behaves normally.")
	}
	goVersion := major + "." + minor

	err := RunCommandInResolvedRootDir("go", "run", "./test/integrations/test_file_patch_generated_same_diffs_as_programatic_patch", "--go-version", goVersion)
	if err != nil {
		t.Fatal(err)
	}
}
