package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "ae27ae104ac2f782ec47dff97216958f65e5723a+1"
const NUMBER = 235

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
