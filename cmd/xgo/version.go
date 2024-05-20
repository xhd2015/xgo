package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "cec0b300e2a4c0d5d0137936c2e68a67ea1798f7+1"
const NUMBER = 224

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
