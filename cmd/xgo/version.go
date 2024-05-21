package main

import "fmt"

const VERSION = "1.0.36"
const REVISION = "110daef2be989ffe0f7a2111e4a8e75272a4b6d3+1"
const NUMBER = 227

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
