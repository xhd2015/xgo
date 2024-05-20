package test_explorer

import (
	"encoding/json"
	"fmt"
)

type TestConfig struct {
	Go      *GoConfig
	GoCmd   string
	Exclude []string
	Env     map[string]interface{}
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
			for _, x := range e {
				s, ok := x.(string)
				if !ok {
					return nil, fmt.Errorf("exclude requires string, actual: %T", x)
				}
				conf.Exclude = append(conf.Exclude, s)
			}
		default:
			return nil, fmt.Errorf("exclude requires string or list, actual: %T", e)
		}
	}

	return conf, nil
}
