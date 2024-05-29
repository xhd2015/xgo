package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "da25b0b8838244b76b23707349c5a2b343abc5d9+1"
const NUMBER = 242

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
