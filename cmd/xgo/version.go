package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "26299095b0e070cff5e5daa2d6e59cea598d1b35+1"
const NUMBER = 183

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
