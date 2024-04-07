package main

import "fmt"

const VERSION = "1.0.19"
const REVISION = "55018f6d640a578b197c4a03220e7a3122326f62+1"
const NUMBER = 168

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
