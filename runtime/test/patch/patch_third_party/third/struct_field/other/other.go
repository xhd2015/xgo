package other

type Other struct {
}

func (c *Other) Greet(name string) string {
	return "hello " + name
}
