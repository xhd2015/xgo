# xgo
Enable function trap for `go`, and provide tools like trace, mock to help go developers write unit test and debug faster.

`xgo` is the successor of the original [go-mock](https://github.com/xhd2015/go-mock).

# Install
```sh
go install github.com/xhd2015/xgo/cmd/xgo
```

# Usage
NOTE: current `xgo` requires at least `go1.16` to compile.
```sh
xgo run ./test/testdata/hello_world
# output:
#  hello world
```

`xgo` works as a drop-in replacement for `go run`,`go build`, and `go test`.