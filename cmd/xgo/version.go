package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "2e709f7b1620aea78fdc1dba282bcf6778dd6d51+1"
const NUMBER = 245

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
