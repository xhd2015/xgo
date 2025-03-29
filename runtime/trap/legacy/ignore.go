package legacy

import (
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
)

var ignoreMap sync.Map // funcinfo -> bool

// mark functions that should skip trap
func Ignore(f interface{}) {
	if f == nil {
		return
	}
	_, funcInfo := Inspect(f)
	if funcInfo == nil {
		return
	}
	ignoreMap.Store(funcInfo, true)
}

// assume f is not nil

func funcIgnored(f *core.FuncInfo) bool {
	_, ok := ignoreMap.Load(f)
	return ok
}
