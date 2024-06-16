package main

import "fmt"

const VERSION = "1.0.40"
const REVISION = "1a16a856a24d88d7f84f61a39340a2b264e99b10+1"
const NUMBER = 266

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
