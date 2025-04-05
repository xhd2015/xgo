//go:build ignore
// +build ignore

package trap_with_overlay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const overlayFile = "overlay.json"

func TestPreCheck(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		t.Fatal(err)
	}
	replace := map[string]string{
		filepath.Join(wd, "trap_with_overlay_test.go"): filepath.Join(wd, "replace_test.go.txt"),
	}
	overlay := map[string]interface{}{
		"Replace": replace,
	}
	overlayJSON, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(overlayFile, overlayJSON, 0755)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostCheck(t *testing.T) {
	err := os.RemoveAll(overlayFile)
	if err != nil {
		t.Fatal(err)
	}
}
