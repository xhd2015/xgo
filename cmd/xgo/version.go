package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "c84c8a6b5fa49005a741ff078078c16d317b1709+1"
const NUMBER = 195

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
