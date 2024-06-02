package main

import "fmt"

const VERSION = "1.0.38"
const REVISION = "a72902358db8db4c239550d226a8360a26ffb904+1"
const NUMBER = 257

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
