package patch

import (
	"context"
	"strings"
)

func greet(s string) string {
	return "hello " + s
}

func greetVaradic(s ...string) string {
	return "hello " + strings.Join(s, ",")
}

func (c *struct_) greetCtx(ctx context.Context) string {
	return "hello " + c.s
}
