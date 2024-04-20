package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "7148c433a32baac7170639d32d6eac241b62bd40+1"
const NUMBER = 192

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
