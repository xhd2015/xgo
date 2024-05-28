package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "5d0b62062accb5c87ec7643e925d351ce65e3b59+1"
const NUMBER = 241

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
