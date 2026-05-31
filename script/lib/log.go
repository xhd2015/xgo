package lib

import (
	"fmt"
	"os"
)

func Logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[cleanup] "+format+"\n", args...)
}

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
