//go:build go1.24

package as_unit_test_run

import (
	"runtime"
	"strings"
	"testing"
)

func TestFilePatchCanBeRepeatedOnPatchedGoroot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: requires bash (make.bash)")
	}
	goVersion, err := goMinorVersion()
	if err != nil {
		t.Error(err)
		return
	}

	args := []string{"run", "./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot", "--go-version", goVersion}
	if strings.Contains(runtime.Version(), "devel") {
		// There is no downloadable release for a development toolchain yet.
		// Test the actual source snapshot instead of resolving its minor version.
		args = append(args, "--goroot", runtime.GOROOT())
	}
	err = RunCommandInResolvedRootDir("go", args...)
	if err != nil {
		t.Fatal(err)
	}
}
