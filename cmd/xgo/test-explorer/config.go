package test_explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/goinfo"
)

type TestConfig struct {
	Go      *GoConfig              `json:"go"`
	GoCmd   string                 `json:"go_cmd"`
	Exclude []string               `json:"exclude"`
	Env     map[string]interface{} `json:"env"`

	// test flags passed to go test
	// common usages:
	//   -p=12            parallel programs
	//   -parallel=12     parallel test cases within the same test
	// according to our test, -p is more useful than -parallel
	Flags []string `json:"flags"`
	Args  []string `json:"args"`
}

func (c *TestConfig) CmdEnv() []string {
	if c == nil || len(c.Env) == 0 {
		return nil
	}
	var env []string
	for k, v := range c.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, fmt.Sprint(v)))
	}
	return env
}

func (c *TestConfig) GetGoCmd() string {
	if c.GoCmd != "" {
		return c.GoCmd
	}
	return "go"
}

type GoConfig struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

func parseTestConfig(config string) (*TestConfig, error) {
	if config == "" {
		return &TestConfig{}, nil
	}
	var m map[string]interface{}
	err := json.Unmarshal([]byte(config), &m)
	if err != nil {
		return nil, err
	}

	conf := &TestConfig{}

	e, ok := m["env"]
	if ok {
		e, ok := e.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("env type err, expect map[string]interface{}, actual: %T", e)
		}
		conf.Env = e
	}

	e, ok = m["go"]
	if ok {
		goConf := &GoConfig{}
		if s, ok := e.(string); ok {
			goConf.Min = s
		} else {
			edata, err := json.Marshal(e)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(edata, &goConf)
			if err != nil {
				return nil, err
			}
		}
		conf.Go = goConf
	}
	e, ok = m["go_cmd"]
	if ok {
		if s, ok := e.(string); ok {
			conf.GoCmd = s
		} else {
			return nil, fmt.Errorf("go_cmd requires string, actual: %T", e)
		}
	}
	e, ok = m["exclude"]
	if ok {
		switch e := e.(type) {
		case string:
			if e != "" {
				conf.Exclude = []string{e}
			}
		case []interface{}:
			list, err := toStringList(e)
			if err != nil {
				return nil, fmt.Errorf("exclude: %w", err)
			}
			conf.Exclude = list
		default:
			return nil, fmt.Errorf("exclude requires string or list, actual: %T", e)
		}
	}
	e, ok = m["flags"]
	if ok {
		list, err := toStringList(e)
		if err != nil {
			return nil, fmt.Errorf("flags: %w", err)
		}
		conf.Flags = list
	}
	e, ok = m["args"]
	if ok {
		list, err := toStringList(e)
		if err != nil {
			return nil, fmt.Errorf("args: %w", err)
		}
		conf.Args = list
	}

	return conf, nil
}

func parseConfigAndMergeOptions(configFile string, opts *Options, configFileRequired bool) (*TestConfig, error) {
	var data []byte
	if configFile != "" {
		var readErr error
		data, readErr = ioutil.ReadFile(configFile)
		if readErr != nil {
			if configFileRequired || !errors.Is(readErr, os.ErrNotExist) {
				return nil, readErr
			}
			readErr = nil
		}
	}
	var testConfig *TestConfig
	if len(data) > 0 {
		var err error
		testConfig, err = parseTestConfig(string(data))
		if err != nil {
			return nil, fmt.Errorf("parse test.config.json: %w", err)
		}
	}
	if testConfig == nil {
		testConfig = &TestConfig{}
	}
	if opts.GoCommand != "" {
		testConfig.GoCmd = opts.GoCommand
	} else if testConfig.GoCmd == "" {
		testConfig.GoCmd = opts.DefaultGoCommand
	}
	testConfig.Exclude = append(testConfig.Exclude, opts.Exclude...)
	testConfig.Flags = append(testConfig.Flags, opts.Flags...)
	testConfig.Args = append(testConfig.Args, opts.Args...)
	return testConfig, nil
}

func validateGoVersion(testConfig *TestConfig) error {
	if testConfig == nil || testConfig.Go == nil || (testConfig.Go.Min == "" && testConfig.Go.Max == "") {
		return nil
	}
	// check go version
	goVersionStr, err := goinfo.GetGoVersionOutput("go")
	if err != nil {
		return err
	}
	goVersion, err := goinfo.ParseGoVersion(goVersionStr)
	if err != nil {
		return err
	}
	if testConfig.Go.Min != "" {
		minVer, _ := goinfo.ParseGoVersionNumber(strings.TrimPrefix(testConfig.Go.Min, "go"))
		if minVer != nil {
			if compareGoVersion(goVersion, minVer, true) < 0 {
				return fmt.Errorf("go version %s < %s", strings.TrimPrefix(goVersionStr, "go version "), testConfig.Go.Min)
			}
		}
	}
	if testConfig.Go.Max != "" {
		maxVer, _ := goinfo.ParseGoVersionNumber(strings.TrimPrefix(testConfig.Go.Max, "go"))
		if maxVer != nil {
			if compareGoVersion(goVersion, maxVer, true) > 0 {
				return fmt.Errorf("go version %s > %s", strings.TrimPrefix(goVersionStr, "go version "), testConfig.Go.Max)
			}
		}
	}
	return nil
}

func parseConfigAndValidate(configFile string, opts *Options, configFileRequired bool) error {
	testConfig, err := parseConfigAndMergeOptions(configFile, opts, configFileRequired)
	if err != nil {
		return err
	}
	return validateGoVersion(testConfig)
}

func toStringList(e interface{}) ([]string, error) {
	if e == nil {
		return nil, nil
	}
	list, ok := e.([]interface{})
	if !ok {
		return nil, fmt.Errorf("requires []string, actual: %T", e)
	}
	strList := make([]string, 0, len(list))
	for _, x := range list {
		s, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("elements requires string, actual: %T", x)
		}
		strList = append(strList, s)
	}
	return strList, nil
}
