package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "53b18b05f3c84c33887d7373607183daff63ce03+1"
const NUMBER = 216

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
