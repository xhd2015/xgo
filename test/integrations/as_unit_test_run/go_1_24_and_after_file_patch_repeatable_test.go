//go:build go1.24

package as_unit_test_run

import (
	"testing"
)

func TestFilePatchCanBeRepeatedOnPatchedGoroot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	goVersion, err := goMinorVersion()
	if err != nil {
		t.Error(err)
		return
	}

	err = RunCommandInResolvedRootDir("go", "run", "./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot", "--go-version", goVersion)
	if err != nil {
		t.Fatal(err)
	}
}
