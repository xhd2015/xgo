package main

import "fmt"

const VERSION = "1.0.34"
const REVISION = "d73b24dafc105fe4610e9130b0db429892961b80+1"
const NUMBER = 215

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
