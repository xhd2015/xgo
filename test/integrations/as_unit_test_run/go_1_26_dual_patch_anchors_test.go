package as_unit_test_run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGo126FileBasedPatchesContainRequiredAnchors(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// as_unit_test_run lives under test/integrations/as_unit_test_run
	repoRoot := filepath.Clean(filepath.Join(root, "..", "..", ".."))

	checks := []struct {
		rel    string
		must   []string
	}{
		{
			rel: "cmd/xgo/asset/patches/go1.26/src/cmd/go/internal/work/exec.go.xgo.patch",
			must: []string{"runCover", "__xgo_overlay_source_file"},
		},
		{
			rel: "cmd/xgo/asset/patches/go1.26/src/cmd/go/internal/test/test.go.xgo.patch",
			must: []string{"moduleLoaderState", "xgoModuleLoaderState"},
		},
		{
			rel: "patches/go1.26/CHANGELOG",
			must: []string{"omniaura/go1.26-support", "go1.27+"},
		},
	}
	for _, c := range checks {
		path := filepath.Join(repoRoot, c.rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", c.rel, err)
		}
		for _, needle := range c.must {
			if !strings.Contains(string(data), needle) {
				t.Fatalf("%s: missing %q", c.rel, needle)
			}
		}
	}
}