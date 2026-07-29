module github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/with_flag

go 1.18

require (
	github.com/xhd2015/xgo/runtime v0.0.0
	github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/app v0.0.0
	github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/other v0.0.0
)

replace github.com/xhd2015/xgo/runtime => ../../../../

replace github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/app => ../app

replace github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/lib => ../lib

replace github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/other => ../other
