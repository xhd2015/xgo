package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

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

type FileOptions struct {
	FilterRules []Rule `json:"filter_rules"`
}

func mergeOptionFiles(tmpDir string, optionFromFile string, mockRules []string) (newFile string, content []byte, err error) {
	if len(mockRules) == 0 {
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

	newOptionFile, err := json.Marshal(FileOptions{FilterRules: mergedRules})
	if err != nil {
		return "", nil, err
	}
	newFile = filepath.Join(tmpDir, "options-from-file.json")
	err = fileutil.WriteFile(newFile, newOptionFile)
	return newFile, newOptionFile, err
}
