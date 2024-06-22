package main

import "fmt"

const VERSION = "1.0.40"
const REVISION = "cb5d3025fad60e00be0dfe8fdc7eb1bef97dedc6+1"
const NUMBER = 267

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
