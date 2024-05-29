package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "f52fb37283a2f6c877d110627da71df63638a916+1"
const NUMBER = 243

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
