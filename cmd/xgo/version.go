package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "62c6c037c1f57c371e227dd1bb8b8e141367f1c6+1"
const NUMBER = 239

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
