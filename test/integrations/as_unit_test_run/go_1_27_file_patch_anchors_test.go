package as_unit_test_run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGo127FileBasedPatchesContainRequiredAnchors(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// as_unit_test_run lives under test/integrations/as_unit_test_run
	repoRoot := filepath.Clean(filepath.Join(root, "..", "..", ".."))

	checks := []struct {
		rel  string
		must []string
	}{
		{
			rel:  "cmd/xgo/asset/patches/go1.27/src/runtime/proc.go.xgo.patch",
			must: []string{"close(mainInitDoneChan)", "__xgo_callback_on_init_finished"},
		},
		{
			rel:  "cmd/xgo/asset/patches/go1.27/src/cmd/go/internal/test/test.go.xgo.patch",
			must: []string{"moduleLoader", "xgoModuleLoader = moduleLoader"},
		},
		{
			// .go files are stored as .go.txt in embedded assets
			rel:  "cmd/xgo/asset/patches/go1.27/src/cmd/go/internal/test/xgo_testunified.go.txt",
			must: []string{"xgoModuleLoader *modload.Loader", "PackagesAndErrors(xgoModuleLoader,"},
		},
		{
			rel:  "patches/go1.27/CHANGELOG",
			must: []string{"go1.27rc2", "mainInitDoneChan", "modload.Loader"},
		},
		{
			rel:  "patches/go1.27/__config__.json",
			must: []string{`"version": "go1.27+"`},
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
