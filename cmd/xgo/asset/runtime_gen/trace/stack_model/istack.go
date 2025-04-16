package stack_model

type IStack interface {
	Data() *Stack
	JSON() ([]byte, error)
}
