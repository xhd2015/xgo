package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	patches "github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

func ApplyFileBased(rootDir, goroot string, goVersion *goinfo.GoVersion) error {
	srcDir := filepath.Join(rootDir, "patches", fmt.Sprintf("go%d.%d", goVersion.Major, goVersion.Minor))
	tmpPatchDir, err := os.MkdirTemp("", "xgo-patch-test-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpPatchDir)
	RunLogged("", nil, "cp", "-R", srcDir+"/", tmpPatchDir+"/")

	config := map[string]interface{}{
		"version":  fmt.Sprintf("go%d.%d", goVersion.Major, goVersion.Minor),
		"copy":     JsonCopyEntries(srcDir),
		"generate": []interface{}{},
	}
	configBytes, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(tmpPatchDir, "__config__.json"), configBytes, 0644)

	extraEnv := map[string]string{
		"XGO_SRC":            rootDir,
		"INSTRUMENT_GOROOT":  goroot,
		"ORIG_GOROOT":        goroot,
		"GO_VERSION":         fmt.Sprintf("go%d.%d.%d", goVersion.Major, goVersion.Minor, goVersion.Patch),
		"GOOS":               runtime.GOOS,
		"GOARCH":             runtime.GOARCH,
	}
	return patches.ApplyPatches(tmpPatchDir, goroot, rootDir, extraEnv, nil, nil)
}

func JsonCopyEntries(srcDir string) interface{} {
	configPath := filepath.Join(srcDir, "__config__.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return []interface{}{}
	}
	var cfg struct {
		Copy json.RawMessage `json:"copy"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return []interface{}{}
	}
	if len(cfg.Copy) > 0 && cfg.Copy[0] == '[' {
		var entries []map[string]interface{}
		json.Unmarshal(cfg.Copy, &entries)
		return entries
	}
	return []interface{}{}
}
