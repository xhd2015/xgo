package hook

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/runtime/legacy"
)

func __xgo_link_on_init_finished(f func()) {
	if !legacy.V1_0_0 {
		return
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_init_finished(requires xgo).")
}

func __xgo_link_init_finished() bool {
	if !legacy.V1_0_0 {
		return false
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_init_finished(requires xgo).")
	return false
}

func OnInitFinished(f func()) {
	__xgo_link_on_init_finished(f)
}

func InitFinished() bool {
	return __xgo_link_init_finished()
}
