package main

import "fmt"

// auto updated
const VERSION = "1.0.40"
const REVISION = "9d7dea9e679bdba690e2659d809b737582ead220+1"
const NUMBER = 269

// manually updated
const CORE_VERSION = "1.0.40"
const CORE_REVISION = "7b0e45276bdca03c2d25c9a4f507d15e627c9664+1"
const CORE_NUMBER = 265

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
