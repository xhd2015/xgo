package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "808e011ba0498b1aaff762095543c880301dc26b+1"
const NUMBER = 249

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
