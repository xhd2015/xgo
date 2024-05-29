package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "8663e050aab31ade5fb04767b355cbdd830fd926+1"
const NUMBER = 245

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
