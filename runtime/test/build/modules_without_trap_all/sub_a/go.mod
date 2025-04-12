module github.com/xhd2015/xgo/runtime/test/build/modules_without_trap_all/sub_a

go 1.18

require github.com/xhd2015/xgo/runtime/test/build/modules_without_trap_all/sub_b v0.0.0

replace github.com/xhd2015/xgo/runtime/test/build/modules_without_trap_all/sub_b => ../sub_b
