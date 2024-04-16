package main

import "fmt"

const VERSION = "1.0.25"
const REVISION = "6eeff524a454a6032a2294b47d1dc1c892b5545f+1"
const NUMBER = 186

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
