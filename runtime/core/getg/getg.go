package getg

import (
	"errors"
	"unsafe"
)

const __XGO_SKIP_TRAP = true

func Curg() unsafe.Pointer {
	return __xgo_link_getcurg()
}

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	panic(errors.New("xgo failed to link __xgo_link_getcurg"))
}
