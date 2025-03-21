package signal

type StartXgoTraceConfig struct {
	OutputFile string
}

// the request is only for recording purepose
func StartXgoTrace(config StartXgoTraceConfig, request interface{}, fn func() (interface{}, error)) (interface{}, error) {
	return fn()
}
