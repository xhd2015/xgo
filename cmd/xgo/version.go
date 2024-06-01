package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "f8e03d3647811f7d4244cbb68369f9e3401e53b9+1"
const NUMBER = 250

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
