package pkg

import (
	"github.com/xhd2015/xgo/runtime/test/hook/id"
)

var ID int

func init() {
	ID = id.Next()
}
