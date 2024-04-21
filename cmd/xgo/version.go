package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "6a13d77b507fde52f56e8fa8db12e89b2a325a5e+1"
const NUMBER = 197

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
