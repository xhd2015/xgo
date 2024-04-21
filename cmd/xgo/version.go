package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "4e07bbae9c2c6c3dc7acf471b763a7fed0fcd66d+1"
const NUMBER = 196

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
