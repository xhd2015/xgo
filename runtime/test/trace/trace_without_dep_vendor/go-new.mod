module demo

go 1.21

require (
	github.com/xhd2015/xgo/runtime v0.0.0-00010101000000-000000000000
	golang.org/x/example/hello v0.0.0-20240205180059-32022caedd6a
)

replace golang.org/x/example/hello => /Users/xhd2015/Projects/xhd2015/xgo/runtime/test/trace_without_dep_vendor/vendor/golang.org/x/example/hello

replace github.com/xhd2015/xgo/runtime => /tmp/xgo_1025_d19b96030c13286e9410580ddec787e7d37cfec21_BUILD_194/runtime
