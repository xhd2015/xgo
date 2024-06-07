package main

import "fmt"

const VERSION = "1.0.40"
const REVISION = "7b0e45276bdca03c2d25c9a4f507d15e627c9664+1"
const NUMBER = 265

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
