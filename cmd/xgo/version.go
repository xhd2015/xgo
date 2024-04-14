package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "b0dd1873bc3e5e5464f55c7b01a077712dc4c818+1"
const NUMBER = 180

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
