package hook

import (
	"github.com/xhd2015/xgo/runtime/test/hook/id"
	_ "github.com/xhd2015/xgo/runtime/test/hook/pkg"
)

var mainInitID int

func init() {
	mainInitID = id.Next()
}
