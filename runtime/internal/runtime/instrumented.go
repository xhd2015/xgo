package runtime

// NotInstrumentedMessage explains how to run a binary that needs xgo's runtime.
const NotInstrumentedMessage = "xgo: this binary was not built by xgo (for example, plain `go test`); xgo patching/interception requires the binary to be built by xgo: run `go tool xgo test ...` or `xgo test ...`; see https://github.com/xhd2015/xgo/blob/master/doc/ERR_NOT_INSTRUMENTED.md"

func RequireInstrumented() {
	if !Instrumented() {
		panic(NotInstrumentedMessage)
	}
}
