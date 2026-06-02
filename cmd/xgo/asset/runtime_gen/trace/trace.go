package trace

import (
	"github.com/xhd2015/xgo/runtime/core"
	_ "github.com/xhd2015/xgo/runtime/internal/trap"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

// NOTE: don't add more functions to this file,
// it is specially instrumented by xgo compiler
// to call trap automatically

type Config struct {
	// OnFinish is called when the trace is finished
	OnFinish func(stack stack_model.IStack) `json:"-"`
	// OutputFile specifies the file to save the trace
	// in json format, which can be open by:
	//      xgo tool trace <OutputFile>
	OutputFile string `json:"OutputFile,omitempty"`

	// FilterTrace is called to filter the trace
	FilterTrace func(funcInfo *core.FuncInfo) bool `json:"-"`
}

// the `request` and `response` are only for recording purpose
func Trace(config Config, request interface{}, fn func() (interface{}, error)) (response interface{}, err error) {
	return fn()
}
