package main

import "fmt"

const VERSION = "1.0.33"
const REVISION = "045b9e4ba45b4ef1e74d71cf64ffcff98761eaef+1"
const NUMBER = 214

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
