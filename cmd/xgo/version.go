package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "4e6a5615d778b8909e3315a2ead323822581dd0e+1"
const NUMBER = 198

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
