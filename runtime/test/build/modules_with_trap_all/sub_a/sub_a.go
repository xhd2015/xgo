package sub_a

import "github.com/xhd2015/xgo/runtime/test/build/modules_with_trap_all/sub_b"

func SubA() {
	sub_b.SubB()
}
