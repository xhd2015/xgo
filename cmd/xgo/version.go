package main

import "fmt"

// auto updated
const VERSION = "1.0.45"
const REVISION = "dda3d723508dbe0490317e7af9ffba15de82b778+1"
const NUMBER = 293

// manually updated
const CORE_VERSION = "1.0.43"
const CORE_REVISION = "f1cf6698521d5b43da06f012ac3ba5afb1308d27+1"
const CORE_NUMBER = 280

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
