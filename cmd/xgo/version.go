package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "4e124e7cc78b77dd490a81bc8db6d8ebdc7c7837+1"
const NUMBER = 236

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
