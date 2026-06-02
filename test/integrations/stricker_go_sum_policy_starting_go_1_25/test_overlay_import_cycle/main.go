package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	coreGo := filepath.Join(dir, "core", "core.go")
	overlayCoreGo := filepath.Join(dir, "_overlay", "core", "core.go")
	overlayPath := filepath.Join(dir, "overlay.json")

	overlay := map[string]map[string]string{
		"Replace": {
			coreGo: overlayCoreGo,
		},
	}
	data, err := json.Marshal(overlay)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal overlay: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(overlayPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write overlay: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "test", "-overlay", overlayPath, "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
