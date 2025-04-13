module github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/service

go 1.18

require (
	github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/third v0.0.0
	github.com/xhd2015/xgo/runtime v0.0.0
)

replace github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_func/third => ../third

replace github.com/xhd2015/xgo/runtime => ../../../..
