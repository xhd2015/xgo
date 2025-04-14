module github.com/xhd2015/xgo/runtime/test/build/overlay_build_cache_error_with_go

go 1.18

require golang.org/x/example/hello v0.0.0-20250407153444-dd59d6852dfb

require github.com/xhd2015/xgo/runtime/test v0.0.0

require github.com/xhd2015/xgo v1.1.0 // indirect

replace github.com/xhd2015/xgo/runtime/test => ../../
