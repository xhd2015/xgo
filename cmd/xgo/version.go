package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "1d56b338a2c930297d1285877665201ebc0e1077+1"
const NUMBER = 246

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
