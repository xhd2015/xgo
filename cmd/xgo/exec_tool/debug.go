package exec_tool

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func getDebugEnv(xgoCompilerEnableEnv string) map[string]string {
	return map[string]string{
		"COMPILER_ALLOW_IR_REWRITE":      "true",
		"COMPILER_ALLOW_SYNTAX_REWRITE":  "true",
		"COMPILER_DEBUG_IR_REWRITE_FUNC": os.Getenv("COMPILER_DEBUG_IR_REWRITE_FUNC"),
		"COMPILER_DEBUG_IR_DUMP_FUNCS":   os.Getenv("COMPILER_DEBUG_IR_DUMP_FUNCS"),
		XGO_DEBUG_DUMP_IR:                os.Getenv(XGO_DEBUG_DUMP_IR),
		XGO_DEBUG_DUMP_IR_FILE:           os.Getenv(XGO_DEBUG_DUMP_IR_FILE),
		XGO_DEBUG_DUMP_AST:               os.Getenv(XGO_DEBUG_DUMP_AST),
		XGO_DEBUG_DUMP_AST_FILE:          os.Getenv(XGO_DEBUG_DUMP_AST_FILE),
		"GOCACHE":                        os.Getenv("GOCACHE"),
		"GOROOT":                         "../..",
		"PATH":                           "../../bin:${env:PATH}",
		"XGO_COMPILER_ENABLE":            xgoCompilerEnableEnv,
	}
}

func getVscodeDebugCmd(cmd string, xgoCompilerEnableEnv string, args []string) *VscodeDebugConfig {
	return &VscodeDebugConfig{
		Name:    fmt.Sprintf("Launch %s", cmd),
		Type:    "go",
		Request: "launch",
		Mode:    "exec",
		Program: cmd,
		Args:    args,
		Env:     getDebugEnv(xgoCompilerEnableEnv),
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
	m, err := config.ToMap()
	if err != nil {
		return err
	}

	return patchJSONPretty(vscodeLaunchFile, func(launchConfig *VscodeLaunchConfigMap) error {
		var foundIdx int = -1
		for i, exConf := range launchConfig.Configurations {
			if fmt.Sprint(exConf["name"]) == config.Name {
				foundIdx = i
				break
			}
		}
		if foundIdx >= 0 {
			launchConfig.Configurations[foundIdx] = m
		} else {
			launchConfig.Configurations = append(launchConfig.Configurations, m)
		}
		return nil
	})
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

const vscodeRemoteDebug = `{
  "name": "dlv remoe localhost:2345",
  "type": "go",
  "request": "attach",
  "mode": "remote",
  "remotePath": "./",
  "port": 2345,
  "host": "127.0.0.1"
}`
