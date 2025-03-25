package patch

type struct_ struct {
	s string
}

func (c *struct_) greet() string {
	return "hello " + c.s
}
