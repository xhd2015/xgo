package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "1223dc619d8076d507f3bd4e2acf923b772729b4+1"
const NUMBER = 160

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
