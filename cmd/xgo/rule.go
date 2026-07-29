package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
)

// TODO: use generate to ensure options sync
//  see patch/match/match.go,  patch/ctxt/match_options.go

type Rule struct {
	Any        bool    `json:"any"`
	Kind       *string `json:"kind"`
	Pkg        *string `json:"pkg"`
	Name       *string `json:"name"`
	Stdlib     *bool   `json:"stdlib"`
	MainModule *bool   `json:"main_module"`
	Generic    *bool   `json:"generic"`
	Exported   *bool   `json:"exported"`
	Closure    *bool   `json:"closure"`
	Action     string  `json:"action"` // include,exclude or empty
}

// FileOptions is written to options-from-file.json for the instrumented compiler
// (XGO_COMPILER_OPTIONS_FILE). Keep in sync with patch/legacy/ctxt.Options.
type FileOptions struct {
	FilterRules []Rule `json:"filter_rules"`
	// MockRuleIncludeAsMainModule is additive module paths treated as main for
	// mock_rules main_module matching only (not process XGO_MAIN_MODULE).
	MockRuleIncludeAsMainModule []string `json:"mock_rule_include_as_main_module,omitempty"`
}

// parseModuleList splits a comma-separated module list (CLI/env form).
func parseModuleList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func mergeOptionFiles(tmpDir string, optionFromFile string, mockRules []string, includeAsMainModules []string) (newFile string, content []byte, err error) {
	if len(mockRules) == 0 && len(includeAsMainModules) == 0 {
		if optionFromFile != "" {
			content, err = fileutil.ReadFile(optionFromFile)
		}
		return optionFromFile, content, err
	}
	var opts FileOptions
	if optionFromFile != "" {
		optionFromFileContent, err := fileutil.ReadFile(optionFromFile)
		if err != nil {
			return "", nil, err
		}

		if len(optionFromFileContent) > 0 {
			err := json.Unmarshal(optionFromFileContent, &opts)
			if err != nil {
				return "", nil, fmt.Errorf("parse %s: %w", optionFromFile, err)
			}
		}
	}

	rulesFromFiles := opts.FilterRules
	mergedRules := make([]Rule, 0, len(rulesFromFiles)+len(mockRules))
	for _, mockRule := range mockRules {
		if mockRule == "" {
			continue
		}
		var rule Rule
		err := json.Unmarshal([]byte(mockRule), &rule)
		if err != nil {
			return "", nil, fmt.Errorf("parse mock rule: %s %w", mockRule, err)
		}
		mergedRules = append(mergedRules, rule)
	}

	mergedRules = append(mergedRules, rulesFromFiles...)
	opts.FilterRules = mergedRules
	if len(includeAsMainModules) > 0 {
		opts.MockRuleIncludeAsMainModule = includeAsMainModules
	}

	newOptionFile, err := json.Marshal(opts)
	if err != nil {
		return "", nil, err
	}
	newFile = filepath.Join(tmpDir, "options-from-file.json")
	err = fileutil.WriteFile(newFile, newOptionFile)
	return newFile, newOptionFile, err
}
