package main

import "fmt"

const VERSION = "1.0.27"
const REVISION = "6ab4a77013b73241451df46df0614dfaceb37a52+1"
const NUMBER = 201

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
