package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xhd2015/xgo/instrument/instrument_compiler"
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
		"generate": JsonGenerateEntries(srcDir),
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
	patchDir := filepath.Join(tmpPatchDir, fmt.Sprintf("go%d.%d", goVersion.Major, goVersion.Minor))
	skipKinds := skipNonMkbuiltinKinds(JsonGenerateKinds(srcDir))
	return patches.ApplyPatches(patchDir, goroot, rootDir, extraEnv, skipKinds, generateHandler)
}

func skipNonMkbuiltinKinds(allKinds []string) []string {
	var skip []string
	for _, k := range allKinds {
		if k != "mkbuiltin" {
			skip = append(skip, k)
		}
	}
	return skip
}

func generateHandler(kind string, extraEnv map[string]string) error {
	switch kind {
	case "mkbuiltin":
		goVersion, err := goinfo.ParseGoVersion("go version " + extraEnv["GO_VERSION"] + " darwin/amd64")
		if err != nil {
			return fmt.Errorf("parse go version: %w", err)
		}
		return instrument_compiler.MkBuiltin(extraEnv["ORIG_GOROOT"], extraEnv["INSTRUMENT_GOROOT"], goVersion, instrument_compiler.RuntimeExtraDef)
	default:
		return fmt.Errorf("unknown generate kind: %q", kind)
	}
}

func JsonGenerateEntries(srcDir string) interface{} {
	configPath := filepath.Join(srcDir, "__config__.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return []interface{}{}
	}
	var cfg struct {
		Generate json.RawMessage `json:"generate"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return []interface{}{}
	}
	if len(cfg.Generate) > 0 && cfg.Generate[0] == '[' {
		var entries []interface{}
		json.Unmarshal(cfg.Generate, &entries)
		return entries
	}
	return []interface{}{}
}

func JsonGenerateKinds(srcDir string) []string {
	entries := JsonGenerateEntries(srcDir)
	entriesList, ok := entries.([]interface{})
	if !ok {
		return nil
	}
	var kinds []string
	for _, e := range entriesList {
		m, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if kind, ok := m["kind"].(string); ok && kind != "" {
			kinds = append(kinds, kind)
		}
	}
	return kinds
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
