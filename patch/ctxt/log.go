package ctxt

import (
	"fmt"
	"os"
	"time"
)

var dummy = func() {}

func LogSpan(msg string) func() {
	if !XGO_COMPILER_LOG_COST {
		return dummy
	}
	start := time.Now()
	return func() {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, time.Since(start))
	}
}
