package trace

import (
	_ "github.com/xhd2015/xgo/runtime/internal/trap"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

// NOTE: don't add more functions to this file,
// it is specially instrumented by xgo compiler
// to call trap automatically

type Config struct {
	OnFinish   func(stack stack_model.IStack) `json:"-"`
	OutputFile string                         `json:"OutputFile,omitempty"`
}

// the `request` and `response` are only for recording purpose
func Trace(config Config, request interface{}, fn func() (interface{}, error)) (response interface{}, err error) {
	return fn()
}
