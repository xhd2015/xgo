module github.com/xhd2015/xgo/runtime/test/patch/patch_var/generic_return/service

go 1.18

require (
	github.com/xhd2015/xgo/runtime v0.0.0
	github.com/xhd2015/xgo/runtime/test/patch/patch_var/generic_return/third v0.0.0
)

replace github.com/xhd2015/xgo/runtime/test/patch/patch_var/generic_return/third => ../third

replace github.com/xhd2015/xgo/runtime => ../../../../..
