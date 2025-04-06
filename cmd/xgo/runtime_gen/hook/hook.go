package hook

import "github.com/xhd2015/xgo/runtime/internal/runtime"

func OnInitFinished(f func()) {
	runtime.XgoOnInitFinished(f)
}

func InitFinished() bool {
	return runtime.XgoInitFinished()
}
