package hook

import (
	"fmt"
	"os"
)

func __xgo_link_on_init_finished(f func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_init_finished(requires xgo).")
}

func __xgo_link_init_finished() bool {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_init_finished(requires xgo).")
	return false
}

func OnInitFinished(f func()) {
	__xgo_link_on_init_finished(f)
}

func InitFinished() bool {
	return __xgo_link_init_finished()
}
