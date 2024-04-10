package main

import "fmt"

const VERSION = "1.0.21"
const REVISION = "60d0e314753d95e7d630a307ef991ac7bc21807e+1"
const NUMBER = 173

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
