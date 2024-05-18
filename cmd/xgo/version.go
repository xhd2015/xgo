package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "61cf57819a9977020a457c10b44c8aa881984fd8+1"
const NUMBER = 219

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
