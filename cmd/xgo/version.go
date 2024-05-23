package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "f948d832e81c7e70b5cc0fc282f582d7a1069e9f+1"
const NUMBER = 232

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
