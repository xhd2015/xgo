package main

import "fmt"

const VERSION = "1.0.26"
const REVISION = "d19c85ac922d26f26038dc99ae3049de0d91d2f6+1"
const NUMBER = 199

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
