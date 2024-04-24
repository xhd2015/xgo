package main

import "fmt"

const VERSION = "1.0.28"
const REVISION = "2eaf6cf94cd17d2888d6708cb76d194b334c8957+1"
const NUMBER = 202

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
