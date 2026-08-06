package test_explorer

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/testconfig"
)

// TestConfig is the explorer view of shared test.config.json.
type TestConfig = testconfig.Config

// XgoConfig is an alias for the shared type.
type XgoConfig = testconfig.XgoConfig

// CoverageConfig is an alias for the shared type.
type CoverageConfig = testconfig.CoverageConfig

// GoConfig is an alias for the shared type.
type GoConfig = testconfig.GoConfig

func parseTestConfig(config string) (*TestConfig, error) {
	if config == "" {
		return &TestConfig{}, nil
	}
	return testconfig.Parse([]byte(config))
}

func parseConfigAndMergeOptions(configFile string, opts *Options, configFileRequired bool) (*TestConfig, error) {
	var data []byte
	if configFile != "" {
		var readErr error
		data, readErr = fileutil.ReadFile(configFile)
		if readErr != nil {
			if configFileRequired || !errors.Is(readErr, os.ErrNotExist) {
				return nil, readErr
			}
			readErr = nil
		}
	}
	var conf *TestConfig
	if len(data) > 0 {
		var err error
		conf, err = parseTestConfig(string(data))
		if err != nil {
			return nil, fmt.Errorf("parse test.config.json: %w", err)
		}
	}
	if conf == nil {
		conf = &TestConfig{}
	}
	var goCmd string
	if opts.GoCommand != "" {
		goCmd = opts.GoCommand
	} else if conf.GoCmd != "" {
		goCmd = conf.GoCmd
	} else {
		goCmd = opts.DefaultGoCommand
	}
	conf.GoCmd = goCmd
	conf.Exclude = append(conf.Exclude, opts.Exclude...)
	conf.Flags = append(conf.Flags, opts.Flags...)
	if goCmd == "xgo" && len(conf.MockRules) > 0 && getXgoSupportsMockRule() {
		for _, mockRule := range conf.MockRules {
			conf.Flags = append(conf.Flags, "--mock-rule", mockRule)
		}
	}
	conf.Args = append(conf.Args, opts.Args...)

	if opts.Coverage == "false" {
		if conf.Coverage == nil {
			conf.Coverage = &CoverageConfig{Disabled: true}
		} else {
			conf.Coverage.Disabled = true
		}
	} else {
		if conf.Coverage == nil {
			conf.Coverage = &CoverageConfig{}
		}
		if opts.CoverageProfile != "" {
			conf.Coverage.Profile = opts.CoverageProfile
		}
		if opts.CoverageDiffWith != "" {
			conf.Coverage.DiffWith = opts.CoverageDiffWith
		}
	}
	return conf, nil
}

// check if xgo version > 1.0.44
func getXgoSupportsMockRule() bool {
	xgoVersion, err := cmd.Output("xgo", "version")
	if err != nil {
		return false
	}
	// not 1.0., so must after 1.0.44
	if !strings.HasPrefix(xgoVersion, "1.0.") {
		return true
	}
	last := strings.TrimPrefix(xgoVersion, "1.0.")
	lastNum, err := strconv.ParseInt(last, 10, 64)
	if err != nil {
		return false
	}
	return lastNum > 44
}

func validateGoVersion(conf *TestConfig) error {
	return testconfig.ValidateGoVersion(conf, "go")
}

func parseConfigAndValidate(configFile string, opts *Options, configFileRequired bool) error {
	testConfig, err := parseConfigAndMergeOptions(configFile, opts, configFileRequired)
	if err != nil {
		return err
	}
	return validateGoVersion(testConfig)
}
