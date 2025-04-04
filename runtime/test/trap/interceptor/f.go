package interceptor

import (
	"fmt"
	"io"
)

func f(recurseBuf io.Writer) {
	fmt.Fprintf(recurseBuf, "call_f\n")
}
