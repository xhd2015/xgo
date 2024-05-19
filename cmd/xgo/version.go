package main

import "fmt"

const VERSION = "1.0.35"
const REVISION = "0a25bf226a233cd13ddadb1bbd98f9066a46f39e+1"
const NUMBER = 222

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
