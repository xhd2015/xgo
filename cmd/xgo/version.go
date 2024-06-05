package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "7fa581a5041d839180502f1d2377ea043803bde7+1"
const NUMBER = 262

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
