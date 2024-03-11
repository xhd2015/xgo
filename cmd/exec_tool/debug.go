package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getVscodeDebugCmd(cmd string, args []string) *VscodeDebugConfig {
	return &VscodeDebugConfig{
		Name:    fmt.Sprintf("Launch %s", cmd),
		Type:    "go",
		Request: "launch",
		Mode:    "exec",
		Program: cmd,
		Args:    args,
		Env: map[string]string{
			"COMPILER_ALLOW_IR_REWRITE":      "true",
			"COMPILER_ALLOW_SYNTAX_REWRITE":  "true",
			"COMPILER_DEBUG_IR_REWRITE_FUNC": os.Getenv("COMPILER_DEBUG_IR_REWRITE_FUNC"),
			"COMPILER_DEBUG_IR_DUMP_FUNCS":   os.Getenv("COMPILER_DEBUG_IR_DUMP_FUNCS"),
			"GOCACHE":                        os.Getenv("GOCACHE"),
			"GOROOT":                         "../..",
			"PATH":                           "../../bin:${env:PATH}",
		},
	}
}

type VscodeDebugConfig struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Request string            `json:"request"`
	Mode    string            `json:"mode"`
	Program string            `json:"program"`
	Cwd     string            `json:"cwd"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type VscodeLaunchConfig struct {
	Configurations []*VscodeDebugConfig `json:"configurations"`
}

type VscodeLaunchConfigMap struct {
	Configurations []map[string]interface{} `json:"configurations"`
}

func addVscodeDebug(vscodeLaunchFile string, config *VscodeDebugConfig) error {
	dir := filepath.Dir(vscodeLaunchFile)
	if dir == "" {
		return fmt.Errorf("invalid dir")
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(vscodeLaunchFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	var launchConfig VscodeLaunchConfigMap
	if len(data) > 0 {
		err := json.Unmarshal(data, &launchConfig)
		if err != nil {
			return fmt.Errorf("bad launch config: %s: %w", vscodeLaunchFile, err)
		}
	}
	m, err := config.ToMap()
	if err != nil {
		return err
	}

	var found bool
	for i, exConf := range launchConfig.Configurations {
		if fmt.Sprint(exConf["name"]) == config.Name {
			launchConfig.Configurations[i] = m
			found = true
			break
		}
	}
	if !found {
		launchConfig.Configurations = append(launchConfig.Configurations, m)
	}

	newConf, err := json.MarshalIndent(launchConfig, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(vscodeLaunchFile, newConf, 0755)
}

func (c *VscodeDebugConfig) ToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
