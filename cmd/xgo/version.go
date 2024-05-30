package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "862f2c74eab90e376e84f8c2575e34c165c08e40+1"
const NUMBER = 253

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
