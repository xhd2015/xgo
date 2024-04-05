package main

import "fmt"

const VERSION = "1.0.18"
const REVISION = "03d82b3e31832e5947c5d3a7ef8752f4f39db28c+1"
const NUMBER = 162

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
