package signal

type StartXgoTraceConfig struct {
	OutputFile string
}

func StartXgoTrace(config StartXgoTraceConfig, fn func() error) error {
	return fn()
}
