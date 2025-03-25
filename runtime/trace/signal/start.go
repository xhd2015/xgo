package signal

import _ "github.com/xhd2015/xgo/runtime/trace/trace_runtime"

type StartXgoTraceConfig struct {
	OutputFile string `json:"OutputFile,omitempty"`
}

// the `request` and `response` are only for recording purpose
func StartXgoTrace(config StartXgoTraceConfig, request interface{}, fn func() (interface{}, error)) (response interface{}, err error) {
	return fn()
}
