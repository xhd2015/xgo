// Package testconfig parses and validates project test.config.json shared by
// xgo test-explorer and consumers such as doctest.
package testconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/goinfo"
)

// DefaultFileName is the conventional project-root config file name.
const DefaultFileName = "test.config.json"

// Config is the shared subset of test.config.json fields.
// Product-specific keys (e.g. doctest pre_test) may appear in the file and
// are ignored by Parse.
type Config struct {
	Go      *GoConfig              `json:"go"`
	GoCmd   string                 `json:"go_cmd"`
	Exclude []string               `json:"exclude"`
	Env     map[string]interface{} `json:"env"`

	// Flags are passed to go/xgo test (e.g. -p=12, --trap-stdlib=false).
	Flags []string `json:"flags"`
	// Args are test-binary program args (after -args).
	Args []string `json:"args"`

	BypassGoFlags bool `json:"bypass_go_flags"`

	// MockRules are re-marshaled JSON objects for --mock-rule.
	MockRules []string        `json:"mock_rules"`
	Xgo       *XgoConfig      `json:"xgo,omitempty"`
	Coverage  *CoverageConfig `json:"coverage,omitempty"`
}

// GoConfig is the go.min / go.max constraint block.
type GoConfig struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

// XgoConfig holds xgo-specific options from test.config.json.
type XgoConfig struct {
	AutoUpdate bool `json:"auto_update"`
}

// CoverageConfig controls coverage integration (enabled by default when present).
type CoverageConfig struct {
	Disabled bool     `json:"disabled"`
	DiffWith string   `json:"diff_with"`
	Profile  string   `json:"profile"`
	Include  []string `json:"include"`
	Exclude  []string `json:"exclude"`
}

// EnvPairs returns KEY=value strings for child processes (stable key order not guaranteed).
func (c *Config) EnvPairs() []string {
	if c == nil || len(c.Env) == 0 {
		return nil
	}
	var env []string
	for k, v := range c.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, fmt.Sprint(v)))
	}
	return env
}

// CmdEnv is an alias for EnvPairs (historical xgo test-explorer name).
func (c *Config) CmdEnv() []string {
	return c.EnvPairs()
}

// GetGoCmd returns configured go_cmd or "go".
func (c *Config) GetGoCmd() string {
	if c != nil && c.GoCmd != "" {
		return c.GoCmd
	}
	return "go"
}

// Load reads path. Missing file returns (nil, nil). Empty file returns empty Config.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return &Config{}, nil
	}
	return Parse(data)
}

// Parse unmarshals test.config.json bytes into Config.
// Unknown top-level keys (e.g. pre_test) are ignored.
func Parse(data []byte) (*Config, error) {
	if len(data) == 0 {
		return &Config{}, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	conf := &Config{}

	if e, ok := m["env"]; ok && e != nil {
		em, ok := e.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("env type err, expect map[string]interface{}, actual: %T", e)
		}
		conf.Env = em
	}

	if e, ok := m["go"]; ok && e != nil {
		goConf := &GoConfig{}
		if s, ok := e.(string); ok {
			goConf.Min = s
		} else {
			edata, err := json.Marshal(e)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(edata, goConf); err != nil {
				return nil, err
			}
		}
		conf.Go = goConf
	}

	if e, ok := m["go_cmd"]; ok && e != nil {
		s, ok := e.(string)
		if !ok {
			return nil, fmt.Errorf("go_cmd requires string, actual: %T", e)
		}
		conf.GoCmd = s
	}

	if e, ok := m["exclude"]; ok && e != nil {
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

	if e, ok := m["flags"]; ok && e != nil {
		list, err := toStringList(e)
		if err != nil {
			return nil, fmt.Errorf("flags: %w", err)
		}
		conf.Flags = list
	}

	if e, ok := m["args"]; ok && e != nil {
		list, err := toStringList(e)
		if err != nil {
			return nil, fmt.Errorf("args: %w", err)
		}
		conf.Args = list
	}

	if e, ok := m["bypass_go_flags"]; ok && e != nil {
		b, err := toBoolean(e)
		if err != nil {
			return nil, fmt.Errorf("bypass_go_flags: %w", err)
		}
		conf.BypassGoFlags = b
	}

	if e, ok := m["mock_rules"]; ok && e != nil {
		list, err := toMarshaledStrings(e)
		if err != nil {
			return nil, fmt.Errorf("mock_rules: %w", err)
		}
		conf.MockRules = list
	}

	if e, ok := m["xgo"]; ok && e != nil {
		if err := copyViaJSON(e, &conf.Xgo); err != nil {
			return nil, fmt.Errorf("xgo: %w", err)
		}
	}

	if e, ok := m["coverage"]; ok && e != nil {
		if b, ok := e.(bool); ok {
			if !b {
				conf.Coverage = &CoverageConfig{Disabled: true}
			}
		} else {
			if err := copyViaJSON(e, &conf.Coverage); err != nil {
				return nil, fmt.Errorf("coverage: %w", err)
			}
		}
	}

	return conf, nil
}

// ValidateGoVersion checks cfg.Go min/max against the host Go toolchain.
// goBinary defaults to "go" when empty. No-op when cfg or constraints are empty.
// Invalid min/max tokens are ignored (same as historical xgo explorer behavior).
func ValidateGoVersion(cfg *Config, goBinary string) error {
	if cfg == nil {
		return nil
	}
	return ValidateGoConstraint(cfg.Go, goBinary)
}

// ValidateGoConstraint checks a go.min/go.max block against goBinary version.
func ValidateGoConstraint(goCfg *GoConfig, goBinary string) error {
	if goCfg == nil || (goCfg.Min == "" && goCfg.Max == "") {
		return nil
	}
	if strings.TrimSpace(goBinary) == "" {
		goBinary = "go"
	}
	goVersionStr, err := goinfo.GetGoVersionOutput(goBinary)
	if err != nil {
		return err
	}
	goVersion, err := goinfo.ParseGoVersion(goVersionStr)
	if err != nil {
		return err
	}
	display := strings.TrimPrefix(goVersionStr, "go version ")
	if goCfg.Min != "" {
		minVer, _ := goinfo.ParseGoVersionNumber(strings.TrimPrefix(goCfg.Min, "go"))
		if minVer != nil {
			if CompareGoVersion(goVersion, minVer, true) < 0 {
				return fmt.Errorf("go version %s < %s", display, goCfg.Min)
			}
		}
	}
	if goCfg.Max != "" {
		maxVer, _ := goinfo.ParseGoVersionNumber(strings.TrimPrefix(goCfg.Max, "go"))
		if maxVer != nil {
			if CompareGoVersion(goVersion, maxVer, true) > 0 {
				return fmt.Errorf("go version %s > %s", display, goCfg.Max)
			}
		}
	}
	return nil
}

// CompareGoVersion compares major/minor (and patch unless ignorePatch).
// Returns a-b style: negative if a < b, zero if equal, positive if a > b.
func CompareGoVersion(a, b *goinfo.GoVersion, ignorePatch bool) int {
	if a == nil || b == nil {
		return 0
	}
	if a.Major != b.Major {
		return a.Major - b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}
	if ignorePatch {
		return 0
	}
	return a.Patch - b.Patch
}

func copyViaJSON(src interface{}, dst interface{}) error {
	if src == nil {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func toStringList(e interface{}) ([]string, error) {
	if e == nil {
		return nil, nil
	}
	// Accept already []string via JSON round-trip path: only []interface{} from map parse.
	list, ok := e.([]interface{})
	if !ok {
		// Single string is handled by callers for exclude; flags/args use toStringList on arrays only.
		if s, ok := e.(string); ok {
			if s == "" {
				return nil, nil
			}
			return []string{s}, nil
		}
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

func toBoolean(e interface{}) (bool, error) {
	if e == nil {
		return false, nil
	}
	if b, ok := e.(bool); ok {
		return b, nil
	}
	if s, ok := e.(string); ok {
		if s == "true" {
			return true, nil
		}
		if s == "false" {
			// Preserve historical xgo bug/compat: "false" returned true.
			// Actual boolean false is handled above. Callers should use JSON bool.
			return true, nil
		}
	}
	return false, fmt.Errorf("expecting true or false, actual: %v", e)
}

func toMarshaledStrings(e interface{}) ([]string, error) {
	if e == nil {
		return nil, nil
	}
	list, ok := e.([]interface{})
	if !ok {
		return nil, fmt.Errorf("requires []string, actual: %T", e)
	}
	strList := make([]string, 0, len(list))
	for _, x := range list {
		if x == nil {
			continue
		}
		if s, ok := x.(string); ok {
			return nil, fmt.Errorf("elements requires non string, actual: %q", s)
		}
		data, err := json.Marshal(x)
		if err != nil {
			return nil, fmt.Errorf("elements to json failed: %w", err)
		}
		strList = append(strList, string(data))
	}
	return strList, nil
}
