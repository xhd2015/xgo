package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "38ce9874b261e60788217c05946dae60dff3d4fd+1"
const NUMBER = 217

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
