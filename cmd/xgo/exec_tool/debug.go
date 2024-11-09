package exec_tool

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type DebugCompile struct {
	Package  string   `json:"package"`
	Env      []string `json:"env"`
	Compiler string   `json:"compiler"`
	Flags    []string `json:"flags"`
	Files    []string `json:"files"`
}

func getDebugEnvMapping(xgoCompilerEnableEnv string) map[string]string {
	envs := getDebugEnvList(xgoCompilerEnableEnv)
	mapping := make(map[string]string, len(envs))
	for _, env := range envs {
		mapping[env[0]] = env[1]
	}
	return mapping
}

func getDebugEnv(xgoCompilerEnableEnv string) []string {
	envs := getDebugEnvList(xgoCompilerEnableEnv)
	joints := make([]string, 0, len(envs))
	for _, env := range envs {
		joints = append(joints, env[0]+"="+env[1])
	}
	return joints
}

func getDebugEnvList(xgoCompilerEnableEnv string) [][2]string {
	return [][2]string{
		{"XGO_COMPILER_ENABLE", xgoCompilerEnableEnv},
		{"COMPILER_ALLOW_IR_REWRITE", "true"},
		{"COMPILER_ALLOW_SYNTAX_REWRITE", "true"},
		{"COMPILER_DEBUG_IR_REWRITE_FUNC", os.Getenv("COMPILER_DEBUG_IR_REWRITE_FUNC")},
		{"COMPILER_DEBUG_IR_DUMP_FUNCS", os.Getenv("COMPILER_DEBUG_IR_DUMP_FUNCS")},
		{XGO_DEBUG_DUMP_IR, os.Getenv(XGO_DEBUG_DUMP_IR)},
		{XGO_DEBUG_DUMP_IR_FILE, os.Getenv(XGO_DEBUG_DUMP_IR_FILE)},
		{XGO_DEBUG_DUMP_AST, os.Getenv(XGO_DEBUG_DUMP_AST)},
		{XGO_DEBUG_DUMP_AST_FILE, os.Getenv(XGO_DEBUG_DUMP_AST_FILE)},
		{XGO_MAIN_MODULE, os.Getenv(XGO_MAIN_MODULE)},
		{XGO_COMPILE_PKG_DATA_DIR, os.Getenv(XGO_COMPILE_PKG_DATA_DIR)},

		// strace
		{XGO_STACK_TRACE, os.Getenv(XGO_STACK_TRACE)},
		{XGO_STACK_TRACE_DIR, os.Getenv(XGO_STACK_TRACE_DIR)},
		{XGO_STRACE_SNAPSHOT_MAIN_MODULE_DEFAULT, os.Getenv(XGO_STRACE_SNAPSHOT_MAIN_MODULE_DEFAULT)},

		{XGO_STD_LIB_TRAP_DEFAULT_ALLOW, os.Getenv(XGO_STD_LIB_TRAP_DEFAULT_ALLOW)},
		{XGO_DEBUG_COMPILE_PKG, os.Getenv(XGO_DEBUG_COMPILE_PKG)},
		{XGO_DEBUG_COMPILE_LOG_FILE, os.Getenv(XGO_DEBUG_COMPILE_LOG_FILE)},
		{XGO_COMPILER_OPTIONS_FILE, os.Getenv(XGO_COMPILER_OPTIONS_FILE)},
		{XGO_SRC_WD, os.Getenv(XGO_SRC_WD)},
		{"GOROOT", os.Getenv("GOROOT")},
		{"GOPATH", os.Getenv("GOPATH")},
		{"PATH", os.Getenv("PATH")},
		{"GOCACHE", os.Getenv("GOCACHE")},
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
		Env:     getDebugEnvMapping(xgoCompilerEnableEnv),
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
  "name": "dlv remote localhost:2345",
  "type": "go",
  "request": "attach",
  "mode": "remote",
  "remotePath": "./",
  "port": 2345,
  "host": "127.0.0.1"
}`
