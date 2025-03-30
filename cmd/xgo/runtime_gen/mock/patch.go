package mock

import (
	"github.com/xhd2015/xgo/runtime/trap"
)

// Patch replaces `fn` with `replacer` in current goroutine.
// You do not have to manually clean up the replacer, as
// xgo will automatically clear the replacer when
// current gorotuine exits.
// However, if you want to clear the replacer earlier,
// this function returns a clean up function that can be
// used to clear the replacer.
func Patch(fn interface{}, replacer interface{}) func() {
	return trap.PushMockByPatch(fn, replacer)
}
