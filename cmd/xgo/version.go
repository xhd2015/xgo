package main

import "fmt"

const VERSION = "1.0.24"
const REVISION = "56315bee3eb7415a4c3d88d4f26bd79c8333a434+1"
const NUMBER = 185

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
