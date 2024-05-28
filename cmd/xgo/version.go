package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "310d0d44809c8f2ad26761138fb8eb3cc4db75c9+1"
const NUMBER = 238

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
