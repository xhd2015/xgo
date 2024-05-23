package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "ced286540d12f84a48095050174f2a37b376bdcc+1"
const NUMBER = 231

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
