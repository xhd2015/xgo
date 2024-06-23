package main

import "fmt"

// auto updated
const VERSION = "1.0.40"
const REVISION = "1e2b5d70ef300bd4b71c6189cc3ab18058a2d1b8+1"
const NUMBER = 270

// manually updated
const CORE_VERSION = "1.0.40"
const CORE_REVISION = "1e2b5d70ef300bd4b71c6189cc3ab18058a2d1b8+1"
const CORE_NUMBER = 269

func getRevision() string {
	return formatRevision(VERSION, REVISION, NUMBER)
}

func getCoreRevision() string {
	return formatRevision(CORE_VERSION, CORE_REVISION, CORE_NUMBER)
}

func formatRevision(version string, revision string, number int) string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", version, revision, revSuffix, number)
}
