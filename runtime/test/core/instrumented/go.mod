module github.com/xhd2015/xgo/runtime/test/core/instrumented

go 1.18

require (
	example.com/xgo-initprobe v0.0.0
	github.com/xhd2015/xgo/runtime v1.1.0
)

replace github.com/xhd2015/xgo/runtime => ../../../

replace example.com/xgo-initprobe => ./initprobe
