package main

import "fmt"

const VERSION = "1.0.39"
const REVISION = "c35e746c1476c5565030f4505e6c7b69099f922e+1"
const NUMBER = 260

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
