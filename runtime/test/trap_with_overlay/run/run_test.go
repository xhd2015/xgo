package run

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOverlay(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wd = filepath.Dir(wd)
	replace := map[string]string{
		filepath.Join(wd, "trap_with_overlay_test.go"): filepath.Join(wd, "replace_test.go.txt"),
	}
	overlayFile := filepath.Join(wd, "overlay.json")
	overlay := map[string]interface{}{
		"Replace": replace,
	}
	overlayJSON, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(overlayFile, overlayJSON, 0755)
	if err != nil {
		t.Fatal(err)
	}

	projectRoot := wd
	for i := 0; i < 3; i++ {
		projectRoot = filepath.Dir(projectRoot)
	}

	cmd := exec.Command("go", "run", "-tags", "dev", "./cmd/xgo", "test",
		"--project-dir", wd,
		"-overlay", overlayFile,
		"-v",
	)
	cmd.Dir = projectRoot
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
