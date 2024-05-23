package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "63941402b67d43bbac6c7d3eb342916bff70f2a3+1"
const NUMBER = 233

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
